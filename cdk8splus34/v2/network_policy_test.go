package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	kplus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func networkPolicySpec(t *testing.T, chart cdk8s.Chart) map[string]interface{} {
	t.Helper()
	return mapAt(t, manifestOfKind(t, chart, "NetworkPolicy"), "spec")
}

func networkPolicyPod(chart cdk8s.Chart, id string) kplus.Pod {
	return kplus.NewPod(chart, jsii.String(id), &kplus.PodProps{
		Containers: &[]*kplus.ContainerProps{{Image: jsii.String("pod")}},
	})
}

func networkPolicyDeployment(chart cdk8s.Chart, id string) kplus.Deployment {
	return kplus.NewDeployment(chart, jsii.String(id), &kplus.DeploymentProps{
		Containers: &[]*kplus.ContainerProps{{Image: jsii.String("pod")}},
	})
}

func networkPolicyRule(t *testing.T, chart cdk8s.Chart, direction string) map[string]interface{} {
	t.Helper()
	spec := networkPolicySpec(t, chart)
	rules := sliceAt(t, spec, direction)
	if len(rules) != 1 {
		t.Fatalf("%s rules = %d, want 1", direction, len(rules))
	}
	rule, ok := rules[0].(map[string]interface{})
	if !ok {
		t.Fatalf("%s rule has type %T", direction, rules[0])
	}
	ports := sliceAt(t, rule, "ports")
	requireDeepEqual(t, ports, []interface{}{map[string]interface{}{"port": float64(6379), "protocol": "TCP"}})
	return rule
}

func networkPolicyPeer(t *testing.T, rule map[string]interface{}, direction string) []interface{} {
	t.Helper()
	key := "from"
	if direction == "egress" {
		key = "to"
	}
	return sliceAt(t, rule, key)
}

func assertNetworkPolicyPeer(t *testing.T, direction string, makePeer func(cdk8s.Chart) kplus.INetworkPolicyPeer, check func(*testing.T, []interface{})) {
	t.Helper()
	chart := cdk8s.Testing_Chart()
	pod := networkPolicyPod(chart, "Pod")
	policy := kplus.NewNetworkPolicy(chart, jsii.String("Policy"), &kplus.NetworkPolicyProps{Selector: pod})
	peer := makePeer(chart)
	ports := []kplus.NetworkPolicyPort{kplus.NetworkPolicyPort_Tcp(jsii.Number(6379))}
	if direction == "ingress" {
		policy.AddIngressRule(peer, &ports)
	} else {
		policy.AddEgressRule(peer, &ports)
	}
	check(t, networkPolicyPeer(t, networkPolicyRule(t, chart, direction), direction))
}

func mapValue(t *testing.T, value interface{}) map[string]interface{} {
	t.Helper()
	result, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("value has type %T, want map[string]interface{}", value)
	}
	return result
}

func assertSelectedPeer(t *testing.T, peers []interface{}, podLabels, namespaceLabels map[string]interface{}, namespaceNames ...string) {
	t.Helper()
	if len(namespaceNames) == 0 {
		if len(peers) != 1 {
			t.Fatalf("peers = %d, want 1", len(peers))
		}
		peer := mapValue(t, peers[0])
		podSelector := mapAt(t, peer, "podSelector")
		if len(podLabels) == 0 {
			requireDeepEqual(t, podSelector, map[string]interface{}{})
		} else {
			requireDeepEqual(t, podSelector["matchLabels"], podLabels)
		}
		if namespaceLabels != nil {
			requireDeepEqual(t, mapAt(t, peer, "namespaceSelector")["matchLabels"], namespaceLabels)
		}
		return
	}
	if len(peers) != len(namespaceNames) {
		t.Fatalf("peers = %d, want %d", len(peers), len(namespaceNames))
	}
	for index, name := range namespaceNames {
		peer := mapValue(t, peers[index])
		requireDeepEqual(t, mapAt(t, peer, "podSelector")["matchLabels"], podLabels)
		labels := mapAt(t, peer, "namespaceSelector")["matchLabels"].(map[string]interface{})
		for key, value := range namespaceLabels {
			if labels[key] != value {
				t.Fatalf("namespace label %q = %#v, want %#v", key, labels[key], value)
			}
		}
		if labels["kubernetes.io/metadata.name"] != name {
			t.Fatalf("namespace name label = %#v, want %q", labels["kubernetes.io/metadata.name"], name)
		}
	}
}

func TestNetworkPolicy(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L6
	t.Run("IpBlock ipv4", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		except := []*string{jsii.String("172.17.1.0/24")}
		block := kplus.NetworkPolicyIpBlock_Ipv4(chart, jsii.String("Block"), jsii.String("172.17.0.0/16"), &except)
		policy := kplus.NewNetworkPolicy(chart, jsii.String("Policy"), nil)
		policy.AddIngressRule(block, nil)
		peer := networkPolicyPeer(t, mapValue(t, sliceAt(t, networkPolicySpec(t, chart), "ingress")[0]), "ingress")
		requireDeepEqual(t, mapAt(t, mapValue(t, peer[0]), "ipBlock"), map[string]interface{}{
			"cidr": "172.17.0.0/16", "except": []interface{}{"172.17.1.0/24"},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L11
	t.Run("IpBlock throws on invalid ipv4 cidr", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		requirePanicContains(t, "Invalid IPv4 CIDR:", func() {
			kplus.NetworkPolicyIpBlock_Ipv4(chart, jsii.String("Block"), jsii.String("1234"), nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L16
	t.Run("IpBlock ipv6", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		except := []*string{jsii.String("2002::1234:abcd:ffff:c0a8:101/24")}
		block := kplus.NetworkPolicyIpBlock_Ipv6(chart, jsii.String("Block"), jsii.String("2002::1234:abcd:ffff:c0a8:101/64"), &except)
		policy := kplus.NewNetworkPolicy(chart, jsii.String("Policy"), nil)
		policy.AddIngressRule(block, nil)
		peer := networkPolicyPeer(t, mapValue(t, sliceAt(t, networkPolicySpec(t, chart), "ingress")[0]), "ingress")
		requireDeepEqual(t, mapAt(t, mapValue(t, peer[0]), "ipBlock"), map[string]interface{}{
			"cidr": "2002::1234:abcd:ffff:c0a8:101/64", "except": []interface{}{"2002::1234:abcd:ffff:c0a8:101/24"},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L21
	t.Run("IpBlock throws on invalid ipv6 cidr", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		requirePanicContains(t, "Invalid IPv6 CIDR:", func() {
			kplus.NetworkPolicyIpBlock_Ipv6(chart, jsii.String("Block"), jsii.String("1234"), nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L26
	t.Run("IpBlock anyIpv4", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		block := kplus.NetworkPolicyIpBlock_AnyIpv4(chart, jsii.String("Block"))
		policy := kplus.NewNetworkPolicy(chart, jsii.String("Policy"), nil)
		policy.AddIngressRule(block, nil)
		peer := networkPolicyPeer(t, mapValue(t, sliceAt(t, networkPolicySpec(t, chart), "ingress")[0]), "ingress")
		requireDeepEqual(t, mapAt(t, mapValue(t, peer[0]), "ipBlock"), map[string]interface{}{"cidr": "0.0.0.0/0"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L31
	t.Run("IpBlock anyIpv6", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		block := kplus.NetworkPolicyIpBlock_AnyIpv6(chart, jsii.String("Block"))
		policy := kplus.NewNetworkPolicy(chart, jsii.String("Policy"), nil)
		policy.AddIngressRule(block, nil)
		peer := networkPolicyPeer(t, mapValue(t, sliceAt(t, networkPolicySpec(t, chart), "ingress")[0]), "ingress")
		requireDeepEqual(t, mapAt(t, mapValue(t, peer[0]), "ipBlock"), map[string]interface{}{"cidr": "::/0"})
	})

	portManifest := func(t *testing.T, port kplus.NetworkPolicyPort) map[string]interface{} {
		t.Helper()
		chart := cdk8s.Testing_Chart()
		policy := kplus.NewNetworkPolicy(chart, jsii.String("Policy"), nil)
		block := kplus.NetworkPolicyIpBlock_AnyIpv4(chart, jsii.String("Block"))
		ports := []kplus.NetworkPolicyPort{port}
		policy.AddIngressRule(block, &ports)
		return mapValue(t, sliceAt(t, mapValue(t, sliceAt(t, networkPolicySpec(t, chart), "ingress")[0]), "ports")[0])
	}

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L40
	t.Run("NetworkPolicyPort tcp", func(t *testing.T) {
		requireDeepEqual(t, portManifest(t, kplus.NetworkPolicyPort_Tcp(jsii.Number(8080))), map[string]interface{}{"port": float64(8080), "protocol": "TCP"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L44
	t.Run("NetworkPolicyPort tcpRange", func(t *testing.T) {
		requireDeepEqual(t, portManifest(t, kplus.NetworkPolicyPort_TcpRange(jsii.Number(8080), jsii.Number(8085))), map[string]interface{}{"port": float64(8080), "endPort": float64(8085), "protocol": "TCP"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L48
	t.Run("NetworkPolicyPort allTcp", func(t *testing.T) {
		requireDeepEqual(t, portManifest(t, kplus.NetworkPolicyPort_AllTcp()), map[string]interface{}{"port": float64(0), "endPort": float64(65535), "protocol": "TCP"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L52
	t.Run("NetworkPolicyPort udp", func(t *testing.T) {
		requireDeepEqual(t, portManifest(t, kplus.NetworkPolicyPort_Udp(jsii.Number(8080))), map[string]interface{}{"port": float64(8080), "protocol": "UDP"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L56
	t.Run("NetworkPolicyPort udpRange", func(t *testing.T) {
		requireDeepEqual(t, portManifest(t, kplus.NetworkPolicyPort_UdpRange(jsii.Number(8080), jsii.Number(8085))), map[string]interface{}{"port": float64(8080), "endPort": float64(8085), "protocol": "UDP"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L60
	t.Run("NetworkPolicyPort allUcp", func(t *testing.T) {
		requireDeepEqual(t, portManifest(t, kplus.NetworkPolicyPort_AllUdp()), map[string]interface{}{"port": float64(0), "endPort": float64(65535), "protocol": "UDP"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L64
	t.Run("NetworkPolicyPort of", func(t *testing.T) {
		port := kplus.NetworkPolicyPort_Of(&kplus.NetworkPolicyPortProps{Port: jsii.Number(5050), EndPort: jsii.Number(5500), Protocol: kplus.NetworkProtocol_SCTP})
		requireDeepEqual(t, portManifest(t, port), map[string]interface{}{"port": float64(5050), "endPort": float64(5500), "protocol": "SCTP"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L72
	t.Run("can create a policy for a managed pod", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := networkPolicyPod(chart, "Web")
		kplus.NewNetworkPolicy(chart, jsii.String("Policy"), &kplus.NetworkPolicyProps{Selector: pod})
		podLabels := mapAt(t, manifestOfKind(t, chart, "Pod"), "metadata", "labels")
		requireDeepEqual(t, mapAt(t, networkPolicySpec(t, chart), "podSelector", "matchLabels"), podLabels)
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L85
	t.Run("can create a policy for a managed workload resource", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := networkPolicyDeployment(chart, "Web")
		kplus.NewNetworkPolicy(chart, jsii.String("Policy"), &kplus.NetworkPolicyProps{Selector: deployment})
		labels := mapAt(t, manifestOfKind(t, chart, "Deployment"), "spec", "selector", "matchLabels")
		requireDeepEqual(t, mapAt(t, networkPolicySpec(t, chart), "podSelector", "matchLabels"), labels)
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L98
	t.Run("can create a policy for selected pods", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		labels := map[string]*string{"app": jsii.String("web")}
		pods := kplus.Pods_Select(chart, jsii.String("Pods"), &kplus.PodsSelectOptions{Labels: &labels})
		kplus.NewNetworkPolicy(chart, jsii.String("Policy"), &kplus.NetworkPolicyProps{Selector: pods})
		requireDeepEqual(t, mapAt(t, networkPolicySpec(t, chart), "podSelector", "matchLabels"), map[string]interface{}{"app": "web"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L108
	t.Run("can create a policy for all pods", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		kplus.NewNetworkPolicy(chart, jsii.String("Policy1"), nil)
		all := kplus.Pods_All(chart, jsii.String("AllPods"), nil)
		kplus.NewNetworkPolicy(chart, jsii.String("Policy2"), &kplus.NetworkPolicyProps{Selector: all})
		manifests := synth(t, chart)
		count := 0
		for _, value := range manifests {
			if manifest, ok := value.(map[string]interface{}); ok && manifest["kind"] == "NetworkPolicy" {
				count++
				requireDeepEqual(t, mapAt(t, manifest, "spec", "podSelector"), map[string]interface{}{})
			}
		}
		if count != 2 {
			t.Fatalf("network policies = %d, want 2", count)
		}
	})

	trafficCase := func(t *testing.T, direction string, defaultValue kplus.NetworkPolicyTrafficDefault) map[string]interface{} {
		t.Helper()
		chart := cdk8s.Testing_Chart()
		props := &kplus.NetworkPolicyProps{}
		traffic := &kplus.NetworkPolicyTraffic{Default: defaultValue}
		if direction == "ingress" {
			props.Ingress = traffic
		} else {
			props.Egress = traffic
		}
		kplus.NewNetworkPolicy(chart, jsii.String("Policy"), props)
		return networkPolicySpec(t, chart)
	}

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L118
	t.Run("can create a policy that allows all ingress by default", func(t *testing.T) {
		spec := trafficCase(t, "ingress", kplus.NetworkPolicyTrafficDefault_ALLOW)
		requireDeepEqual(t, spec["ingress"], []interface{}{map[string]interface{}{}})
		requireDeepEqual(t, spec["policyTypes"], []interface{}{"Ingress"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L130
	t.Run("can create a policy that denies all ingress by default", func(t *testing.T) {
		spec := trafficCase(t, "ingress", kplus.NetworkPolicyTrafficDefault_DENY)
		if _, exists := spec["ingress"]; exists {
			t.Fatal("deny-all ingress must not synthesize an allow rule")
		}
		requireDeepEqual(t, spec["policyTypes"], []interface{}{"Ingress"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L142
	t.Run("can create a policy that allows all egress by default", func(t *testing.T) {
		spec := trafficCase(t, "egress", kplus.NetworkPolicyTrafficDefault_ALLOW)
		requireDeepEqual(t, spec["egress"], []interface{}{map[string]interface{}{}})
		requireDeepEqual(t, spec["policyTypes"], []interface{}{"Egress"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L154
	t.Run("can create a policy that denies all egress by default", func(t *testing.T) {
		spec := trafficCase(t, "egress", kplus.NetworkPolicyTrafficDefault_DENY)
		if _, exists := spec["egress"]; exists {
			t.Fatal("deny-all egress must not synthesize an allow rule")
		}
		requireDeepEqual(t, spec["policyTypes"], []interface{}{"Egress"})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L166
	t.Run("cannot create a policy for a selector that selects pods in multiple namespaces", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		names := []*string{jsii.String("n1"), jsii.String("n2")}
		namespaces := kplus.Namespaces_Select(chart, jsii.String("Namespaces"), &kplus.NamespacesSelectOptions{Names: &names})
		pods := kplus.Pods_Select(chart, jsii.String("Pods"), &kplus.PodsSelectOptions{Namespaces: namespaces})
		requirePanicContains(t, "selects pods in multiple namespaces", func() {
			kplus.NewNetworkPolicy(chart, jsii.String("Policy"), &kplus.NetworkPolicyProps{Selector: pods})
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L181
	t.Run("cannot create a policy for a selector that selects pods in namespaces based on labels", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		namespaceLabels := map[string]*string{"tier": jsii.String("web")}
		namespaces := kplus.Namespaces_Select(chart, jsii.String("Namespaces"), &kplus.NamespacesSelectOptions{Labels: &namespaceLabels})
		podLabels := map[string]*string{"app": jsii.String("web")}
		pods := kplus.Pods_Select(chart, jsii.String("Pods"), &kplus.PodsSelectOptions{Labels: &podLabels, Namespaces: namespaces})
		requirePanicContains(t, "selects pods in namespaces based on labels", func() {
			kplus.NewNetworkPolicy(chart, jsii.String("Policy"), &kplus.NetworkPolicyProps{Selector: pods})
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L197
	t.Run("policy namespace defaults to selector namespace", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := kplus.NewPod(chart, jsii.String("Web"), &kplus.PodProps{
			Metadata:   &cdk8s.ApiObjectMetadata{Namespace: jsii.String("n1")},
			Containers: &[]*kplus.ContainerProps{{Image: jsii.String("web")}},
		})
		kplus.NewNetworkPolicy(chart, jsii.String("Policy"), &kplus.NetworkPolicyProps{Selector: pod})
		metadata := mapAt(t, manifestOfKind(t, chart, "NetworkPolicy"), "metadata")
		if metadata["namespace"] != "n1" {
			t.Fatalf("namespace = %#v, want n1", metadata["namespace"])
		}
	})

	checkIPBlock := func(t *testing.T, peers []interface{}) {
		t.Helper()
		if len(peers) != 1 {
			t.Fatalf("peers = %d, want 1", len(peers))
		}
		requireDeepEqual(t, mapAt(t, mapValue(t, peers[0]), "ipBlock"), map[string]interface{}{
			"cidr": "172.17.0.0/16", "except": []interface{}{"172.17.1.0/24"},
		})
	}
	makeIPBlock := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer {
		except := []*string{jsii.String("172.17.1.0/24")}
		return kplus.NetworkPolicyIpBlock_Ipv4(chart, jsii.String("Block"), jsii.String("172.17.0.0/16"), &except)
	}
	checkManaged := func(t *testing.T, peers []interface{}) {
		t.Helper()
		if len(peers) != 1 {
			t.Fatalf("peers = %d, want 1", len(peers))
		}
		labels := mapAt(t, mapValue(t, peers[0]), "podSelector", "matchLabels")
		if len(labels) != 1 || labels["cdk8s.io/metadata.addr"] == nil {
			t.Fatalf("managed pod selector labels = %#v", labels)
		}
	}
	makeManagedPod := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer { return networkPolicyPod(chart, "Peer") }
	makeManagedDeployment := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer { return networkPolicyDeployment(chart, "Deployment") }
	makeSelectedPods := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer {
		labels := map[string]*string{"type": jsii.String("selected")}
		return kplus.Pods_Select(chart, jsii.String("Pods"), &kplus.PodsSelectOptions{Labels: &labels})
	}
	checkSelectedPods := func(t *testing.T, peers []interface{}) {
		assertSelectedPeer(t, peers, map[string]interface{}{"type": "selected"}, nil)
	}
	makeSelectedPodsInLabeledNamespaces := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer {
		labels := map[string]*string{"type": jsii.String("selected")}
		namespaces := kplus.Namespaces_Select(chart, jsii.String("Namespaces"), &kplus.NamespacesSelectOptions{Labels: &labels})
		return kplus.Pods_Select(chart, jsii.String("Pods"), &kplus.PodsSelectOptions{Labels: &labels, Namespaces: namespaces})
	}
	checkSelectedPodsInLabeledNamespaces := func(t *testing.T, peers []interface{}) {
		assertSelectedPeer(t, peers, map[string]interface{}{"type": "selected"}, map[string]interface{}{"type": "selected"})
	}
	makeSelectedPodsInNamedNamespaces := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer {
		labels := map[string]*string{"type": jsii.String("selected")}
		names := []*string{jsii.String("selected1"), jsii.String("selected2")}
		namespaces := kplus.Namespaces_Select(chart, jsii.String("Namespaces"), &kplus.NamespacesSelectOptions{Labels: &labels, Names: &names})
		return kplus.Pods_Select(chart, jsii.String("Pods"), &kplus.PodsSelectOptions{Labels: &labels, Namespaces: namespaces})
	}
	checkSelectedPodsInNamedNamespaces := func(t *testing.T, peers []interface{}) {
		assertSelectedPeer(t, peers, map[string]interface{}{"type": "selected"}, map[string]interface{}{"type": "selected"}, "selected1", "selected2")
	}
	makeAllPods := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer {
		return kplus.Pods_All(chart, jsii.String("AllPods"), nil)
	}
	checkAllPods := func(t *testing.T, peers []interface{}) {
		if len(peers) != 1 {
			t.Fatalf("peers = %d, want 1", len(peers))
		}
		requireDeepEqual(t, mapAt(t, mapValue(t, peers[0]), "podSelector"), map[string]interface{}{})
	}
	makeManagedNamespace := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer {
		return kplus.NewNamespace(chart, jsii.String("Namespace"), nil)
	}
	checkManagedNamespace := func(t *testing.T, peers []interface{}) {
		t.Helper()
		if len(peers) != 1 {
			t.Fatalf("peers = %d, want 1", len(peers))
		}
		peer := mapValue(t, peers[0])
		requireDeepEqual(t, mapAt(t, peer, "podSelector"), map[string]interface{}{})
		labels := mapAt(t, peer, "namespaceSelector", "matchLabels")
		if labels["kubernetes.io/metadata.name"] == nil {
			t.Fatalf("managed namespace selector = %#v", labels)
		}
	}
	makeSelectedNamespaces := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer {
		labels := map[string]*string{"type": jsii.String("selected")}
		return kplus.Namespaces_Select(chart, jsii.String("Namespaces"), &kplus.NamespacesSelectOptions{Labels: &labels})
	}
	checkSelectedNamespaces := func(t *testing.T, peers []interface{}) {
		assertSelectedPeer(t, peers, map[string]interface{}{}, map[string]interface{}{"type": "selected"})
	}
	makeAllNamespaces := func(chart cdk8s.Chart) kplus.INetworkPolicyPeer {
		return kplus.Namespaces_All(chart, jsii.String("AllPods"))
	}
	checkAllNamespaces := func(t *testing.T, peers []interface{}) {
		t.Helper()
		if len(peers) != 1 {
			t.Fatalf("peers = %d, want 1", len(peers))
		}
		peer := mapValue(t, peers[0])
		requireDeepEqual(t, mapAt(t, peer, "podSelector"), map[string]interface{}{})
		requireDeepEqual(t, mapAt(t, peer, "namespaceSelector"), map[string]interface{}{})
	}

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L211
	t.Run("can add ingress from an ip block", func(t *testing.T) { assertNetworkPolicyPeer(t, "ingress", makeIPBlock, checkIPBlock) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L226
	t.Run("can add ingress from a managed pod", func(t *testing.T) { assertNetworkPolicyPeer(t, "ingress", makeManagedPod, checkManaged) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L240
	t.Run("can add ingress from a managed workload resource", func(t *testing.T) { assertNetworkPolicyPeer(t, "ingress", makeManagedDeployment, checkManaged) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L254
	t.Run("can add ingress from pods selected without namespaces", func(t *testing.T) { assertNetworkPolicyPeer(t, "ingress", makeSelectedPods, checkSelectedPods) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L269
	t.Run("can add ingress from pods selected with namespaces selected by labes", func(t *testing.T) {
		assertNetworkPolicyPeer(t, "ingress", makeSelectedPodsInLabeledNamespaces, checkSelectedPodsInLabeledNamespaces)
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L287
	t.Run("can add ingress from pods selected with namespaces selected by names", func(t *testing.T) {
		assertNetworkPolicyPeer(t, "ingress", makeSelectedPodsInNamedNamespaces, checkSelectedPodsInNamedNamespaces)
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L308
	t.Run("can add ingress from all pods", func(t *testing.T) { assertNetworkPolicyPeer(t, "ingress", makeAllPods, checkAllPods) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L323
	t.Run("can add ingress from managed namespace", func(t *testing.T) { assertNetworkPolicyPeer(t, "ingress", makeManagedNamespace, checkManagedNamespace) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L338
	t.Run("can add ingress from selected namespaces", func(t *testing.T) {
		assertNetworkPolicyPeer(t, "ingress", makeSelectedNamespaces, checkSelectedNamespaces)
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L353
	t.Run("can add ingress from all namespaces", func(t *testing.T) { assertNetworkPolicyPeer(t, "ingress", makeAllNamespaces, checkAllNamespaces) })

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L369
	t.Run("can add egress to an ip block", func(t *testing.T) { assertNetworkPolicyPeer(t, "egress", makeIPBlock, checkIPBlock) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L384
	t.Run("can add egress to a managed pod", func(t *testing.T) { assertNetworkPolicyPeer(t, "egress", makeManagedPod, checkManaged) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L398
	t.Run("can add egress to a managed workload resource", func(t *testing.T) { assertNetworkPolicyPeer(t, "egress", makeManagedDeployment, checkManaged) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L412
	t.Run("can add egress to pods selected without namespaces", func(t *testing.T) { assertNetworkPolicyPeer(t, "egress", makeSelectedPods, checkSelectedPods) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L427
	t.Run("can add egress to pods selected with namespaces selected by labes", func(t *testing.T) {
		assertNetworkPolicyPeer(t, "egress", makeSelectedPodsInLabeledNamespaces, checkSelectedPodsInLabeledNamespaces)
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L445
	t.Run("can add egress to pods selected with namespaces selected by names", func(t *testing.T) {
		assertNetworkPolicyPeer(t, "egress", makeSelectedPodsInNamedNamespaces, checkSelectedPodsInNamedNamespaces)
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L466
	t.Run("can add egress to all pods", func(t *testing.T) { assertNetworkPolicyPeer(t, "egress", makeAllPods, checkAllPods) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L481
	t.Run("can add egress to managed namespace", func(t *testing.T) { assertNetworkPolicyPeer(t, "egress", makeManagedNamespace, checkManagedNamespace) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L496
	t.Run("can add egress to selected namespaces", func(t *testing.T) {
		assertNetworkPolicyPeer(t, "egress", makeSelectedNamespaces, checkSelectedNamespaces)
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/network-policy.test.ts#L511
	t.Run("can add egress to all namespaces", func(t *testing.T) { assertNetworkPolicyPeer(t, "egress", makeAllNamespaces, checkAllNamespaces) })
}
