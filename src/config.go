package main

type Config struct {
	Device  string   `yaml:"device" required:"true"`
	Rebinds []Rebind `yaml:"rebinds"`
}

type Rebind struct {
	From      string `yaml:"from"`
	To        string `yaml:"to"`
	Unbind    bool   `yaml:"unbind"`
	Modifiers []struct {
		Modifier string `yaml:"modifier"`
		To       string `yaml:"to"`
	} `yaml:"with_modifiers"`
}
