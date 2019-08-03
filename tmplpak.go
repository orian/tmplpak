package tmplpak

import (
	htmlTemplate "html/template"
	"io"
	"net/url"
	"path"
	"sync"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Template interface {
	Execute(wr io.Writer, data interface{}) error

	// ExecuteTemplate applies the template associated with t that has the given
	// Name to the specified data object and writes the output to wr.
	// If an error occurs executing the template or writing its output,
	// execution stops, but partial results may already have been written to
	// the output writer.
	// A template may be executed safely in parallel, although if parallel
	// executions share a Writer the output may be interleaved.
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

type TemplateLoader interface {
	MustGet(name string) Template
	Get(name string) (Template, error)
}

type Loader struct {
	Reload    bool
	LoadFunc  LoadFunc
	BaseDir   string
	templates sync.Map
	configs   map[string]Config
	fmap      FuncMap
}

type LoadFunc func(name string, fmap FuncMap, files ...string) (Template, error)

func TextTemplate(name string, fmap FuncMap, files ...string) (Template, error) {
	if t, err := template.New(name).Funcs(template.FuncMap(fmap)).ParseFiles(files...); err == nil {
		return t.Lookup(name), nil
	} else {
		return nil, err
	}
}

type FuncMap map[string]interface{}

func HtmlTemplate(name string, fmap FuncMap, files ...string) (Template, error) {
	f := path.Base(files[0])
	return htmlTemplate.New(f).Funcs(htmlTemplate.FuncMap(fmap)).ParseFiles(files...)
}

func Load(baseDir string, config Config, fmap FuncMap, lf LoadFunc) (Template, error) {
	var files []string
	for _, v := range config.Files {
		if !path.IsAbs(v) {
			v = path.Join(baseDir, v)
		}
		files = append(files, v)
	}

	return lf(config.Name, fmap, files...)
}

func (m *Loader) load(cfg Config) (Template, error) {
	return m.LoadFunc(m.BaseDir, m.fmap, cfg.Files...)
}

func (m *Loader) get(name string) (Template, error) {
	var err error
	if !m.Reload {
		if v, ok := m.templates.Load(name); ok {
			return v.(Template), nil
		}
	}

	cfg, ok := m.configs[name]
	if !ok {
		return nil, ErrNotFound
	}
	t, err := Load(m.BaseDir, cfg, m.fmap, m.LoadFunc)
	if err != nil {
		return nil, err
	} else if !m.Reload {
		m.templates.Store(name, t)
	}

	return t, err
}

func (m *Loader) MustGet(name string) Template {
	t, err := m.Get(name)
	if err != nil {
		logrus.Panicf("cannot get template: %s", err)
	}
	return t
}

var ErrNotFound = errors.New("not found")

func (m *Loader) Get(name string) (Template, error) {
	return m.get(name)
}

func (m *Loader) Register(config Config) {
	m.configs[config.Name] = config
}

func New(fmap FuncMap, baseDir string) *Loader {
	return &Loader{
		BaseDir:   baseDir,
		LoadFunc:  HtmlTemplate,
		templates: sync.Map{},
		configs:   make(map[string]Config),
		fmap:      fmap,
	}
}

func ResUrl(base string) func(res, s string) string {
	u, _ := url.Parse(base)
	return func(res, s string) string {
		newU := *u
		newU.Path = path.Join(newU.Path, res, s)
		return newU.String()
	}
}

type Config struct {
	Name  string
	Files []string
}

type Site struct {
	Title    string
	Keywords []string
}

type ErrorData struct {
	Site    *Site
	Message string
}
