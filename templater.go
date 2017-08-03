package venom

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

// Templater contains templating values on a testsuite
type Templater struct {
	Values map[string]string
}

func newTemplater(values map[string]string) *Templater {
	if values == nil {
		values = make(map[string]string)
	}
	return &Templater{Values: values}
}

// Add add data to templater
func (tmpl *Templater) Add(prefix string, values map[string]string) {
	if tmpl.Values == nil {
		tmpl.Values = make(map[string]string)
	}
	dot := ""
	if prefix != "" {
		dot = "."
	}
	for k, v := range values {
		tmpl.Values[prefix+dot+k] = v
	}
}

//ApplyOnStep executes the template on a test step
func (tmpl *Templater) ApplyOnStep(step *TestStep) error {
	// Using yaml to encode/decode, it generates map[interface{}]interface{} typed data that json does not like
	s, err := yaml.Marshal(step)
	if err != nil {
		return fmt.Errorf("templater> Error while marshaling: %s", err)
	}

	fmt.Println("STEP")
	fmt.Println(step)

	sb, err := tmpl.apply(s)
	if err != nil {
		return err
	}

	var t TestStep
	if err := yaml.Unmarshal([]byte(sb), &t); err != nil {
		return fmt.Errorf("templater> Error while unmarshal: %s, content:%s", err, sb)
	}

	*step = t
	return nil
}

//ApplyOnContext executes the template on a context
func (tmpl *Templater) ApplyOnContext(ctx map[string]interface{}) (map[string]interface{}, error) {
	// Using yaml to encode/decode, it generates map[interface{}]interface{} typed data that json does not like
	s, err := yaml.Marshal(ctx)
	if err != nil {
		return nil, fmt.Errorf("templater> Error while marshaling: %s", err)
	}
	sb, err := tmpl.apply(s)
	if err != nil {
		return nil, err
	}

	var t map[string]interface{}
	if err := yaml.Unmarshal([]byte(sb), &t); err != nil {
		return nil, fmt.Errorf("templater> Error while unmarshal: %s, content:%s", err, sb)
	}

	return t, nil
}

func (tmpl *Templater) apply(in []byte) ([]byte, error) {
	data := map[string]string{}
	input := string(in)

	for k, v := range tmpl.Values {
		kb := strings.Replace(k, ".", "__", -1)
		data[kb] = v
		re := regexp.MustCompile("{{." + k + "(.*)}}")
		for {
			sm := re.FindStringSubmatch(input)
			if len(sm) > 0 {
				input = strings.Replace(input, sm[0], "{{."+kb+sm[1]+"}}", -1)
			} else {
				break
			}
		}
	}

	funcMap := template.FuncMap{
		"extract": func(args ...interface{}) string {
			fmt.Println(args)
			/*
				r, err := regexp.Compile(s2)
				if err != nil {
					return "Error with regex "
				}

				res := r.FindStringSubmatch(arg)
				if len(res) == 0 {
					fmt.Println("no match", arg)
					return arg
				}

				fmt.Println("Extract=", strings.Join(res[1:], " "))

				return strings.Join(res[1:], " ")*/
			return ""
		},
	}

	fmt.Println(input)

	t, err := template.New("templater").Funcs(funcMap).Parse(string(input))
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	if err := t.Execute(&buff, data); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
