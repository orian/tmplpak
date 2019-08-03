# tmplpak

Very minimalistic library helping to work with templates.

Example usage:

```go
tmpl := tmplpak.New(fmap, dirPath)
	tmpl.Reload = hotReload
	tmpl.Register(tmplpak.Config{"main", []string{"base.html", "main.html"}})
	tmpl.Register(tmplpak.Config{"signin", []string{"base.html", "signin.html"}})
	tmpl.Register(tmplpak.Config{"signout", []string{"base.html", "signout.html"}})
```

and rendering:

```
renderer := &tmplpak.TemplateRenderHelper{
    Log:       baseLog,
    Templates: tmpl,
}

data := map[string]string{
    "Title": "Great success",
    "CanonicalURL": "http://example.com/some/path"
}

renderer.Render(w, r, "page-edit-from", data)
```

Hot reloading allow to reload templates during the development phase.
```
tmpl.Reload = true  # Enables hot reloading of templates.
``` 