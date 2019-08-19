package tmplpak

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"math/rand"
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

// ServeErrorTemplate serves an error template. It returns an log entry
// with ErrorID field set. The ErrorID is passed to template it can be presented
// to the end user, so he can refer to it.
func (b *TemplateRenderHelper) ServeErrorTemplate(w http.ResponseWriter, r *http.Request,
	err error, externalMsg string, statusCode int) *logrus.Entry {

	randSlice := make([]byte, 16)
	rand.Read(randSlice)
	errorID := base64.RawURLEncoding.EncodeToString(randSlice)
	e := b.Log.WithError(err).WithField("eid", errorID)

	t, innerErr := b.Templates.Get(b.ErrorTemplate)
	if innerErr != nil {
		if innerErr != ErrNotFound {
			b.Log.WithError(innerErr).Error("when rendering error template")
		}
		b.ServeError(w, r, externalMsg, http.StatusUnauthorized)
		return e
	}
	w.WriteHeader(statusCode)
	data := map[string]interface{}{
		"StatusCode": statusCode,
		"Status":     externalMsg,
		"ErrorID":    errorID,
	}
	if err := t.Execute(w, data); err != nil {
		b.Log.WithError(err).Error("execute template")
	}
	return e
}

// JSON encodes the data in JSON format. It ensures the proper headers are set.
// If no encoder is provided it uses the default json.NewEncoder.
func (b *TemplateRenderHelper) JSON(w http.ResponseWriter, data interface{}) {
	// https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html#rule-31---html-escape-json-values-in-an-html-context-and-read-the-data-with-jsonparse
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	var enc JSONEncoder
	if b.Encoder == nil {
		enc = json.NewEncoder(w)
	} else {
		enc = b.Encoder(w)
	}
	if err := enc.Encode(data); err != nil {
		logrus.WithError(err).Error("encoding value as JSON")
	}
}

func (b *TemplateRenderHelper) Clone() *TemplateRenderHelper {
	return &TemplateRenderHelper{
		Log:           b.Log,
		Templates:     b.Templates,
		ErrorTemplate: b.ErrorTemplate,
		Encoder:       b.Encoder,
	}
}

// JSONEncoder is interface for json.Encoder requiring to implement Encode only.
type JSONEncoder interface {
	Encode(v interface{}) error
}

type JSONEncoderGenerator func(w io.Writer) JSONEncoder

type TemplateRenderHelper struct {
	Log           logrus.FieldLogger
	Templates     TemplateLoader
	ErrorTemplate string
	Encoder       JSONEncoderGenerator
}
