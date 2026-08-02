package cdk8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

type includeImpl struct {
	constructBase
}

func NewInclude(scope constructs.Construct, id *string, props *IncludeProps) Include {
	if scope == nil || id == nil || props == nil || props.Url == nil {
		panic("scope, id, and props.url are required")
	}
	result := &includeImpl{}
	initializeInclude(result, result, scope, id, props)
	return result
}

func NewInclude_Override(include Include, scope constructs.Construct, id *string, props *IncludeProps) {
	if include == nil || scope == nil || id == nil || props == nil || props.Url == nil {
		panic("include, scope, id, and props.url are required")
	}
	result, ok := include.(*includeImpl)
	if ok {
		initializeInclude(result, result, scope, id, props)
		return
	}
	result = &includeImpl{}
	if !setEmbeddedImplementation(include, result) {
		panic("cdk8s: Include override must embed cdk8s.Include")
	}
	initializeInclude(result, include, scope, id, props)
}

func initializeInclude(storage *includeImpl, host Include, scope constructs.Construct, id *string, props *IncludeProps) {
	if host == storage {
		constructs.NewConstruct_Override(storage, scope, id)
	} else {
		storage.node = constructs.NewNode(host, scope, id)
	}
	documents := Yaml_Load(props.Url)
	if documents != nil {
		addIncludedDocuments(host, *documents)
	}
}

func addIncludedDocuments(scope constructs.Construct, documents []interface{}) {
	unnamedOrder := 0
	for _, document := range documents {
		manifest := plainMap(document)
		apiVersion, _ := manifest["apiVersion"].(string)
		kind, _ := manifest["kind"].(string)
		if apiVersion == "" || kind == "" {
			panic("included manifest requires apiVersion and kind")
		}
		metadata, _ := manifest["metadata"].(map[string]interface{})
		name, hasName := metadata["name"].(string)
		if !hasName {
			name = fmt.Sprintf("object%d", unnamedOrder)
			unnamedOrder++
		}
		namespace, _ := metadata["namespace"].(string)
		parts := make([]string, 0, 3)
		if name != "" {
			parts = append(parts, name)
		}
		if kind != "" {
			parts = append(parts, strings.ToLower(kind))
		}
		if namespace != "" {
			parts = append(parts, namespace)
		}
		identifier := strings.Join(parts, "-")
		NewApiObjectWithManifest(scope, jsii.String(identifier), &ApiObjectProps{
			ApiVersion: jsii.String(apiVersion),
			Kind:       jsii.String(kind),
		}, manifest)
	}
}

func (i *includeImpl) ApiObjects() *[]ApiObject {
	return includedObjects(i.Node())
}

func includedObjects(node constructs.Node) *[]ApiObject {
	result := make([]ApiObject, 0)
	for _, child := range *node.Children() {
		if object, ok := child.(ApiObject); ok {
			result = append(result, object)
		}
	}
	return &result
}

// Checks if `x` is a construct.
//
// Returns: true if `x` is an object created from a class which extends `Construct`. Deprecated: use `x instanceof Construct` instead.
func Include_IsConstruct(value interface{}) *bool {
	return App_IsConstruct(value)
}

type helmImpl struct {
	constructBase
	releaseName string
}

const maxHelmBuffer = 10 * 1024 * 1024

var errHelmBufferExceeded = errors.New("helm maxBuffer length exceeded")

type helmOutputBuffer struct {
	buffer   bytes.Buffer
	limit    int
	overflow bool
	cancel   context.CancelFunc
}

func (b *helmOutputBuffer) Write(data []byte) (int, error) {
	if b.buffer.Len()+len(data) <= b.limit {
		return b.buffer.Write(data)
	}
	remaining := b.limit - b.buffer.Len()
	if remaining > 0 {
		_, _ = b.buffer.Write(data[:remaining])
	}
	b.overflow = true
	b.cancel()
	return remaining, errHelmBufferExceeded
}

func NewHelm(scope constructs.Construct, id *string, props *HelmProps) Helm {
	if scope == nil || id == nil || props == nil || props.Chart == nil {
		panic("scope, id, and props.chart are required")
	}
	result := &helmImpl{}
	initializeHelm(result, result, scope, id, props)
	return result
}

func NewHelm_Override(helm Helm, scope constructs.Construct, id *string, props *HelmProps) {
	if helm == nil || scope == nil || id == nil || props == nil || props.Chart == nil {
		panic("helm, scope, id, and props.chart are required")
	}
	result, ok := helm.(*helmImpl)
	if ok {
		initializeHelm(result, result, scope, id, props)
		return
	}
	result = &helmImpl{}
	if !setEmbeddedImplementation(helm, result) {
		panic("cdk8s: Helm override must embed cdk8s.Helm")
	}
	initializeHelm(result, helm, scope, id, props)
}

func initializeHelm(helm *helmImpl, host Helm, scope constructs.Construct, id *string, props *HelmProps) {
	workdir, err := os.MkdirTemp("", "cdk8s-helm-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(workdir)

	args := []string{"template"}
	if props.Values != nil && len(*props.Values) > 0 {
		valuesPath := filepath.Join(workdir, "overrides.yaml")
		content := Yaml_Stringify(*props.Values)
		if err := os.WriteFile(valuesPath, []byte(stringValue(content)), 0o644); err != nil {
			panic(err)
		}
		args = append(args, "-f", valuesPath)
	}
	if props.Repo != nil {
		args = append(args, "--repo", *props.Repo)
	}
	if props.Version != nil {
		args = append(args, "--version", *props.Version)
	}
	if props.Namespace != nil {
		args = append(args, "--namespace", *props.Namespace)
	}
	if props.HelmFlags != nil {
		for _, flag := range *props.HelmFlags {
			if flag != nil {
				args = append(args, *flag)
			}
		}
	}

	releaseName := props.ReleaseName
	if releaseName == nil {
		maxLen := float64(53)
		extra := []*string{id}
		releaseName = Names_ToDnsLabel(scope, &NameOptions{MaxLen: &maxLen, Extra: &extra})
	}
	args = append(args, *releaseName, *props.Chart)

	program := "helm"
	if props.HelmExecutable != nil {
		program = *props.HelmExecutable
	}
	commandContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	command := exec.CommandContext(commandContext, program, args...)
	command.WaitDelay = time.Second
	stdout := &helmOutputBuffer{limit: maxHelmBuffer, cancel: cancel}
	stderr := &helmOutputBuffer{limit: maxHelmBuffer, cancel: cancel}
	command.Stdout = stdout
	command.Stderr = stderr
	err = command.Run()
	if stdout.overflow {
		panic("stdout maxBuffer length exceeded")
	}
	if stderr.overflow {
		panic("stderr maxBuffer length exceeded")
	}
	if errors.Is(err, exec.ErrNotFound) || os.IsNotExist(err) {
		panic(fmt.Sprintf("unable to execute '%s' to render Helm chart. Is it installed on your system?", program))
	}
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			panic(stderr.buffer.String())
		}
		panic(fmt.Sprintf("error while rendering a helm chart: %v", err))
	}
	output := stdout.buffer.Bytes()

	if host == helm {
		constructs.NewConstruct_Override(helm, scope, id)
	} else {
		helm.node = constructs.NewNode(host, scope, id)
	}
	documentsPath := filepath.Join(workdir, "chart.yaml")
	if err := os.WriteFile(documentsPath, output, 0o644); err != nil {
		panic(err)
	}
	documents := Yaml_Load(&documentsPath)
	if documents != nil {
		addIncludedDocuments(host, *documents)
	}
	helm.releaseName = *releaseName
}

func (h *helmImpl) ApiObjects() *[]ApiObject {
	return includedObjects(h.Node())
}

func (h *helmImpl) ReleaseName() *string {
	return &h.releaseName
}

// Checks if `x` is a construct.
//
// Returns: true if `x` is an object created from a class which extends `Construct`. Deprecated: use `x instanceof Construct` instead.
func Helm_IsConstruct(value interface{}) *bool {
	return App_IsConstruct(value)
}
