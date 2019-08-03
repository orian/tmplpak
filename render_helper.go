package tmplpak

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func (b *TemplateRenderHelper) Render(w http.ResponseWriter, r *http.Request, name string, dat interface{}) {
	t, err := b.Templates.Get(name)
	if err != nil {
		b.Log.WithError(err).Errorf("cannot find template: %s", name)
		b.ServeError(w, r, "Please try again later", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, dat); err != nil {
		b.Log.WithError(err).Error("execute template")
	}
}

func (b *TemplateRenderHelper) ServeError(w http.ResponseWriter, r *http.Request,
	externalMsg string, statusCode int) {

	http.Error(w, externalMsg, statusCode)
}

type TemplateRenderHelper struct {
	Log       logrus.FieldLogger
	Templates TemplateLoader
}
