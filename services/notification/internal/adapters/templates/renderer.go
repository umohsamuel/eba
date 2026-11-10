package templates

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"io/fs"
	"maps"
	"path"
	texttemplate "text/template"

	"github.com/umohsamuel/eba/pkg/events"
	domain "github.com/umohsamuel/eba/services/notification/internal/domain/notification"
	service "github.com/umohsamuel/eba/services/notification/internal/services/notification"
)

//go:embed files
var embedded embed.FS

type definition struct {
	subject  string
	html     string
	text     string
	required []string
}

var definitions = map[string]definition{
	events.TemplateAuthWelcome: {
		subject: "Welcome to {{.app_name}}",
		html:    "auth/welcome.html",
		text:    "auth/welcome.txt",
	},
	events.TemplateAuthVerifyEmail: {
		subject:  "Verify your email",
		html:     "auth/verify_email.html",
		text:     "auth/verify_email.txt",
		required: []string{"verify_url", "expires_in"},
	},
	events.TemplateAuthPasswordReset: {
		subject:  "Reset your password",
		html:     "auth/password_reset.html",
		text:     "auth/password_reset.txt",
		required: []string{"reset_url", "expires_in"},
	},
	events.TemplateAuthPasswordChanged: {
		subject: "Your password was changed",
		html:    "auth/password_changed.html",
		text:    "auth/password_changed.txt",
	},
}

type compiled struct {
	subject  *texttemplate.Template
	html     *htmltemplate.Template
	text     *texttemplate.Template
	textName string
	required []string
}

type Renderer struct {
	appName   string
	templates map[string]compiled
}

func New(appName string) (*Renderer, error) {
	files, err := fs.Sub(embedded, "files")
	if err != nil {
		return nil, err
	}

	r := &Renderer{appName: appName, templates: make(map[string]compiled, len(definitions))}

	for name, def := range definitions {
		subject, err := texttemplate.New("subject").Option("missingkey=error").Parse(def.subject)
		if err != nil {
			return nil, fmt.Errorf("parse %s subject: %w", name, err)
		}

		html, err := htmltemplate.New("layout").
			Option("missingkey=error").
			Funcs(htmltemplate.FuncMap{"dict": dict}).
			ParseFS(files, "layout.html", def.html)
		if err != nil {
			return nil, fmt.Errorf("parse %s html: %w", name, err)
		}

		text, err := texttemplate.New(path.Base(def.text)).Option("missingkey=error").ParseFS(files, def.text)
		if err != nil {
			return nil, fmt.Errorf("parse %s text: %w", name, err)
		}

		r.templates[name] = compiled{
			subject:  subject,
			html:     html,
			text:     text,
			textName: path.Base(def.text),
			required: def.required,
		}
	}

	return r, nil
}

func (r *Renderer) Render(name string, data map[string]string) (*service.Rendered, error) {
	t, ok := r.templates[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", domain.ErrUnknownTemplate, name)
	}

	for _, key := range t.required {
		if data[key] == "" {
			return nil, fmt.Errorf("%w: template %s requires %q", domain.ErrInvalidRequest, name, key)
		}
	}

	values := make(map[string]string, len(data)+1)
	maps.Copy(values, data)
	values["app_name"] = r.appName

	var subject, html, text bytes.Buffer

	if err := t.subject.Execute(&subject, values); err != nil {
		return nil, fmt.Errorf("%w: render %s subject: %v", domain.ErrInvalidRequest, name, err)
	}
	if err := t.html.ExecuteTemplate(&html, "layout", values); err != nil {
		return nil, fmt.Errorf("%w: render %s html: %v", domain.ErrInvalidRequest, name, err)
	}
	if err := t.text.ExecuteTemplate(&text, t.textName, values); err != nil {
		return nil, fmt.Errorf("%w: render %s text: %v", domain.ErrInvalidRequest, name, err)
	}

	return &service.Rendered{
		Subject: subject.String(),
		HTML:    html.String(),
		Text:    text.String(),
	}, nil
}

func dict(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, errors.New("dict needs key/value pairs")
	}

	m := make(map[string]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict key %v is not a string", pairs[i])
		}
		m[key] = pairs[i+1]
	}

	return m, nil
}
