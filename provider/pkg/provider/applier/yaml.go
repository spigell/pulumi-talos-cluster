package applier

import (
	"bytes"
	"fmt"
	"maps"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Guard func(root map[string]any) error

type Merger struct {
	root map[string]any // merged head (yaml1 <- yaml2 first doc)
	tail string         // verbatim extra docs from yaml2 (starting with '---')
	err  error
}

func MergeYAML(yaml1, yaml2 string) *Merger {
	m := &Merger{}

	var h1 any
	if err := yaml.Unmarshal([]byte(yaml1), &h1); err != nil {
		m.err = fmt.Errorf("yaml1 parse: %w", err)
		return m
	}
	r1, ok := normalize(h1).(map[string]any)
	if !ok {
		m.err = fmt.Errorf("yaml1 top-level is not a mapping")
		return m
	}

	// Split all yaml2 docs
	docs := splitYaml2All(yaml2)
	var verbatim []string
	docNum := 0

	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}
		docNum++

		var h any
		if err := yaml.Unmarshal([]byte(doc), &h); err != nil {
			m.err = fmt.Errorf("yaml2 doc %d parse: %w", docNum, err)
			return m
		}
		m2, ok := normalize(h).(map[string]any)
		if !ok {
			m.err = fmt.Errorf("yaml2 doc %d top-level is not a mapping", docNum)
			return m
		}

		hasAPIV := false
		hasKind := false
		_, hasAPIV = m2["apiVersion"]
		_, hasKind = m2["kind"]

		switch {
		case hasAPIV && hasKind:
			verbatim = append(verbatim, doc)

		case hasAPIV != hasKind:
			missing := "kind"
			if hasKind {
				missing = "apiVersion"
			}
			m.err = fmt.Errorf("yaml2 doc %d has %s but missing %s", docNum, map[bool]string{hasAPIV: "apiVersion", hasKind: "kind"}[true], missing)
			return m

		default:
			// merge
			r1 = mergeMaps(r1, m2)
		}
	}

	m.root = r1
	if len(verbatim) > 0 {
		var sb strings.Builder
		for _, doc := range verbatim {
			sb.WriteString("---\n")
			sb.WriteString(doc)
			if !strings.HasSuffix(doc, "\n") {
				sb.WriteString("\n")
			}
		}
		m.tail = strings.TrimRight(sb.String(), "\n")
	}

	return m
}

func (m *Merger) WithGuard(g Guard) *Merger {
	if m.err != nil || g == nil {
		return m
	}
	if err := g(m.root); err != nil {
		m.err = err
	}
	return m
}

func (m *Merger) Build() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(m.root); err != nil {
		_ = enc.Close()
		return "", fmt.Errorf("encode merged head: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", fmt.Errorf("close encoder: %w", err)
	}
	out := strings.TrimRight(buf.String(), "\n")
	if strings.TrimSpace(m.tail) != "" {
		if !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		out += m.tail // append yaml2's extra docs verbatim
	}
	return out, nil
}

// ----- Guards -----

// GuardUnmodifyK8sImages overwrites any images in the merged head
// to ensure we don't downgrade Kubernetes components.
func GuardUnmodifyK8sImages(img *K8SImages) Guard {
	return func(root map[string]any) error {
		// machine.kubelet.image
		setPath(root, []string{"machine", "kubelet", "image"}, img.Kubelet)
		// cluster.apiServer.image
		setPath(root, []string{"cluster", "apiServer", "image"}, img.APIServer)
		// cluster.controllerManager.image
		setPath(root, []string{"cluster", "controllerManager", "image"}, img.ControllerManager)
		// cluster.scheduler.image
		setPath(root, []string{"cluster", "scheduler", "image"}, img.Scheduler)
		// cluster.proxy.image
		setPath(root, []string{"cluster", "proxy", "image"}, img.KubeProxy)
		return nil
	}
}

// ----- Internals -----.
func normalize(v any) any {
	switch t := v.(type) {
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, vv := range t {
			m[fmt.Sprint(k)] = normalize(vv)
		}
		return m
	case map[string]any:
		for k, vv := range t {
			t[k] = normalize(vv)
		}
		return t
	case []any:
		for i := range t {
			t[i] = normalize(t[i])
		}
		return t
	default:
		return v
	}
}

func mergeMaps(dst, src map[string]any) map[string]any {
	out := make(map[string]any, len(dst))
	maps.Copy(out, dst)
	for k, v2 := range src {
		if v1, ok := out[k]; ok {
			switch a := v1.(type) {
			case map[string]any:
				if b, ok := v2.(map[string]any); ok {
					out[k] = mergeMaps(a, b)
					continue
				}
			case []any:
				if b, ok := v2.([]any); ok {
					out[k] = append(a, b...) // change to replace if you prefer
					continue
				}
			}
		}
		out[k] = v2
	}
	return out
}

// setPath ensures nested maps exist and sets the terminal value.
func setPath(root map[string]any, path []string, val any) {
	m := root
	for i := 0; i < len(path)-1; i++ {
		k := path[i]
		nxt, ok := m[k].(map[string]any)
		if !ok {
			nxt = map[string]any{}
			m[k] = nxt
		}
		m = nxt
	}
	m[path[len(path)-1]] = val
}

func splitYaml2All(s string) []string {
	re := regexp.MustCompile(`(?m)^\s*---\s*\n`)
	return re.Split(s, -1)
}
