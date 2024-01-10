package general

type LayoutTemplate struct {
	IsAdmin bool
}

func (t *LayoutTemplate) Template() string { return "settings/general/content.tmpl" }
