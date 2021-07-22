# hell-flow

```
helm repo add hellflow https://raw.githubusercontent.com/tamalsaha/hell-flow/master/stable/
helm repo update
```

**--set values**

- https://github.com/helm/helm/blob/main/pkg/cli/values/options.go#L60-L72
- https://github.com/helm/helm/blob/main/pkg/strvals/parser.go#L415-L446

**Render Templates**

- https://github.com/helm/helm/blob/43853ea772e779b617b011a8d4e69205636fa4f9/pkg/engine/engine.go#L190-L261

**Zero Template Chart**

- Must always use Install action
- Can't use Install or Upgrade to remove any YAMLs


** ToDos **

- [x] Auto Register Application CRD


**Multi-chart**

- ownership checks for resources

```
  metadata:
    annotations:
      meta.helm.sh/release-name: first
      meta.helm.sh/release-namespace: default
    labels:
      app.kubernetes.io/managed-by: Helm
```

- storage driver ownership

```
	// apply labels
	lbs.set("name", rls.Name)
	lbs.set("owner", "helm")
	lbs.set("status", rls.Info.Status.String())
	lbs.set("version", strconv.Itoa(rls.Version))
```

** Chart Annotations **

- https://artifacthub.io/docs/topics/annotations/helm/
- https://github.com/helm/helm/blob/v3.5.4/pkg/chart/metadata.go#L71-L73

```
	// Annotations are additional mappings uninterpreted by Helm,
	// made available for inspection by other applications.
	Annotations map[string]string `json:"annotations,omitempty"`
```

- app.kubernetes.io/part-of
- meta.x-helm.dev/editor: {gvr}
- meta.x-helm.dev/resources: |
   - { GK }
   - { GK }
