package types

type ContactForm struct {
	Token                string `json:"token"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Message              string `json:"message"`
	*Mapper[ContactForm] `json:"-" yaml:"-" xml:"-" toml:"-" gorm:"-"`
}
