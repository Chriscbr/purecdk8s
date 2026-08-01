package importer

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type crdManifest struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Items      []*crdManifest `yaml:"items"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec *crdSpec `yaml:"spec"`
}

type crdSpec struct {
	Group string `yaml:"group"`
	Names struct {
		Kind string `yaml:"kind"`
	} `yaml:"names"`
	Version    string `yaml:"version"`
	Validation *struct {
		OpenAPIV3Schema *schema `yaml:"openAPIV3Schema"`
	} `yaml:"validation"`
	Versions []crdVersion `yaml:"versions"`
}

type crdVersion struct {
	Name   string `yaml:"name"`
	Served *bool  `yaml:"served"`
	Schema *struct {
		OpenAPIV3Schema *schema `yaml:"openAPIV3Schema"`
	} `yaml:"schema"`
}

type parsedCRD struct {
	Group    string
	Kind     string
	Versions []parsedCRDVersion
}

type parsedCRDVersion struct {
	Name   string
	Schema *schema
}

// GenerateCRDs generates one Go package per API group found in a Kubernetes
// apiextensions.k8s.io/v1 (or legacy v1beta1) CRD manifest. PackageName, when
// set, is used only when the manifest contains a single group. PackagePrefix
// prefixes every normalized API-group package name.
func GenerateCRDs(data []byte, options GenerateOptions) ([]*Generation, error) {
	crds, err := decodeCRDs(data)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]parsedCRD)
	for _, crd := range crds {
		grouped[crd.Group] = append(grouped[crd.Group], crd)
	}
	groups := sortedKeys(grouped)
	result := make([]*Generation, 0, len(groups))
	for _, group := range groups {
		groupOptions := options
		groupPackage := sanitizePackageName(group)
		switch {
		case groupOptions.PackagePrefix != "":
			groupOptions.PackageName = prefixedPackageName(groupOptions.PackagePrefix, groupPackage)
		case groupOptions.PackageName == "" || len(groups) > 1:
			groupOptions.PackageName = groupPackage
		}
		defs := make(map[string]*schema)
		resources := make([]resourceDefinition, 0)
		items := grouped[group]
		sort.Slice(items, func(i, j int) bool {
			return strings.ToLower(items[i].Kind) < strings.ToLower(items[j].Kind)
		})
		for _, crd := range items {
			for index, version := range crd.Versions {
				suffix := ""
				if index > 0 {
					suffix = versionSuffix(version.Name)
				}
				prefix := groupOptions.ClassNamePrefix
				typeName := resourceTypeName(prefix, crd.Kind, version.Name, true, suffix)
				item := version.Schema
				if item == nil {
					item = &schema{Type: "object", Properties: map[string]*schema{}}
				}
				fqn := crd.Kind + suffix
				defs[fqn] = item
				for name, definition := range item.Definitions {
					defs[name] = definition
				}
				for name, definition := range item.Defs {
					defs[name] = definition
				}
				resources = append(resources, resourceDefinition{
					FQN:       fqn,
					Group:     group,
					Version:   version.Name,
					Kind:      crd.Kind,
					Prefix:    prefix,
					Suffix:    suffix,
					Custom:    true,
					Schema:    item,
					TypeName:  typeName,
					PropsName: normalizeTypeName(typeName + "Props"),
				})
			}
		}
		generated, err := generatePackage(defs, resources, groupOptions, group)
		if err != nil {
			return nil, fmt.Errorf("generate CRD group %s: %w", group, err)
		}
		result = append(result, generated)
	}
	return result, nil
}

func decodeCRDs(data []byte) ([]parsedCRD, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var manifests []*crdManifest
	for {
		var manifest crdManifest
		err := decoder.Decode(&manifest)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode CustomResourceDefinition manifest: %w", err)
		}
		collectCRDManifests(&manifests, &manifest)
	}
	if len(manifests) == 0 {
		return nil, fmt.Errorf("CustomResourceDefinition manifest contains no CRDs")
	}

	byKey := make(map[string]*parsedCRD)
	order := make([]string, 0)
	for _, manifest := range manifests {
		if manifest.APIVersion != "apiextensions.k8s.io/v1" && manifest.APIVersion != "apiextensions.k8s.io/v1beta1" {
			return nil, fmt.Errorf("invalid CustomResourceDefinition %q: unsupported apiVersion %q", manifest.Metadata.Name, manifest.APIVersion)
		}
		if manifest.Spec == nil {
			return nil, fmt.Errorf("invalid CustomResourceDefinition %q: spec is required", manifest.Metadata.Name)
		}
		if manifest.Spec.Group == "" || manifest.Spec.Names.Kind == "" {
			return nil, fmt.Errorf("invalid CustomResourceDefinition %q: spec.group and spec.names.kind are required", manifest.Metadata.Name)
		}
		versions := make([]parsedCRDVersion, 0)
		if manifest.Spec.Version != "" {
			var root *schema
			if manifest.Spec.Validation != nil {
				root = manifest.Spec.Validation.OpenAPIV3Schema
			}
			versions = append(versions, parsedCRDVersion{Name: manifest.Spec.Version, Schema: root})
		} else {
			for _, version := range manifest.Spec.Versions {
				if version.Name == "" {
					return nil, fmt.Errorf("invalid CustomResourceDefinition %q: version name is required", manifest.Metadata.Name)
				}
				var root *schema
				if version.Schema != nil {
					root = version.Schema.OpenAPIV3Schema
				}
				if root == nil && manifest.Spec.Validation != nil {
					root = manifest.Spec.Validation.OpenAPIV3Schema
				}
				versions = append(versions, parsedCRDVersion{Name: version.Name, Schema: root})
			}
		}
		if len(versions) == 0 {
			return nil, fmt.Errorf("invalid CustomResourceDefinition %q: no versions found", manifest.Metadata.Name)
		}
		key := manifest.Spec.Group + "/" + strings.ToLower(manifest.Spec.Names.Kind)
		existing := byKey[key]
		if existing == nil {
			existing = &parsedCRD{Group: manifest.Spec.Group, Kind: manifest.Spec.Names.Kind}
			byKey[key] = existing
			order = append(order, key)
		}
		seen := make(map[string]bool)
		for _, version := range existing.Versions {
			seen[version.Name] = true
		}
		for _, version := range versions {
			if seen[version.Name] {
				return nil, fmt.Errorf("invalid CustomResourceDefinition %q: duplicate version %q", manifest.Metadata.Name, version.Name)
			}
			seen[version.Name] = true
			existing.Versions = append(existing.Versions, version)
		}
	}
	sort.Strings(order)
	result := make([]parsedCRD, 0, len(order))
	for _, key := range order {
		result = append(result, *byKey[key])
	}
	return result, nil
}

func collectCRDManifests(result *[]*crdManifest, manifest *crdManifest) {
	if manifest == nil {
		return
	}
	if manifest.Kind == "CustomResourceDefinition" {
		*result = append(*result, manifest)
	}
	if manifest.Kind == "List" {
		for _, item := range manifest.Items {
			collectCRDManifests(result, item)
		}
	}
}
