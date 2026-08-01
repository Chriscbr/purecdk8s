package main

import (
	"strings"

	"example.com/purecdk8s-integration/imports/k8s"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

type addNamePrefixResolver struct {
	prefix string
}

func (resolver *addNamePrefixResolver) Resolve(context cdk8s.ResolutionContext) {
	key := context.Key()
	if key == nil || len(*key) != 2 || stringPointerValue((*key)[0]) != "metadata" || stringPointerValue((*key)[1]) != "name" {
		return
	}

	value, ok := resolvedString(context.Value())
	if !ok || strings.HasPrefix(value, resolver.prefix) {
		return
	}
	context.ReplaceValue(resolver.prefix + value)
}

func resolvedString(value interface{}) (string, bool) {
	switch value := value.(type) {
	case string:
		return value, true
	case *string:
		if value != nil {
			return *value, true
		}
	}
	return "", false
}

func stringPointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func main() {
	resolver := &addNamePrefixResolver{prefix: "integration-"}
	resolvers := []cdk8s.IResolver{resolver}
	app := cdk8s.NewApp(&cdk8s.AppProps{Resolvers: &resolvers})
	chart := cdk8s.NewChart(app, jsii.String("app"), nil)

	data := map[string]*string{
		"message": jsii.String("hello from cdk8s"),
	}
	k8s.NewKubeConfigMap(chart, jsii.String("settings"), &k8s.KubeConfigMapProps{
		Data: &data,
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("settings"),
		},
	})

	app.Synth()
}
