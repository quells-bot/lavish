package main

import (
	"encoding/binary"
	"hash/fnv"
	"testing"

	"github.com/quells/lavish"
	"github.com/quells/lavish/engine/preact10"
)

// FormData holds the dynamic values passed from Go to the JSX component
type FormData struct {
	FormTitle       string
	InputLabel      string
	InputName       string
	DropdownLabel   string
	DropdownName    string
	DropdownOptions []DropdownOption
}

type DropdownOption struct {
	Value string
	Label string
}

func (d FormData) Hash64() uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(d.FormTitle))
	_, _ = h.Write([]byte(d.InputLabel))
	_, _ = h.Write([]byte(d.InputName))
	_, _ = h.Write([]byte(d.DropdownLabel))
	_, _ = h.Write([]byte(d.DropdownName))
	for _, opt := range d.DropdownOptions {
		_, _ = h.Write([]byte(opt.Value))
		_, _ = h.Write([]byte(opt.Label))
	}
	return binary.BigEndian.Uint64(h.Sum(nil))
}

const formJSX = `
const Form = ({ title, inputLabel, inputName, dropdownLabel, dropdownName, options }) => (
    <form>
        <h2>{title}</h2>
        <div>
            <label for={inputName}>{inputLabel}</label>
            <input type="text" id={inputName} name={inputName} />
        </div>
        <div>
            <label for={dropdownName}>{dropdownLabel}</label>
            <select id={dropdownName} name={dropdownName}>
                {options.map((opt) => (
                    <option value={opt.Value}>{opt.Label}</option>
                ))}
            </select>
        </div>
        <button type="submit">Submit</button>
    </form>
);

const { FormTitle, InputLabel, InputName, DropdownLabel, DropdownName, DropdownOptions } = data;

render(
    <Form
        title={FormTitle}
        inputLabel={InputLabel}
        inputName={InputName}
        dropdownLabel={DropdownLabel}
        dropdownName={DropdownName}
        options={DropdownOptions}
    />
)
`

func TestRenderForm(t *testing.T) {
	renderer := preact10.RenderEngine
	data := FormData{
		FormTitle:     "Contact Us",
		InputLabel:    "Your Name",
		InputName:     "username",
		DropdownLabel: "Select Country",
		DropdownName:  "country",
		DropdownOptions: []DropdownOption{
			{Value: "us", Label: "United States"},
			{Value: "uk", Label: "United Kingdom"},
			{Value: "ca", Label: "Canada"},
		},
	}

	bundle := lavish.NewBundle(renderer)

	got, err := bundle.RenderJSX("form.jsx", formJSX, data)
	if err != nil {
		t.Fatal(err)
	}

	expect := `<form>` +
		`<h2>Contact Us</h2>` +
		`<div>` +
		`<label for="username">Your Name</label>` +
		`<input type="text" id="username" name="username" />` +
		`</div>` +
		`<div>` +
		`<label for="country">Select Country</label>` +
		`<select id="country" name="country">` +
		`<option value="us">United States</option>` +
		`<option value="uk">United Kingdom</option>` +
		`<option value="ca">Canada</option>` +
		`</select>` +
		`</div>` +
		`<button type="submit">Submit</button>` +
		`</form>`

	if got != expect {
		t.Fatalf("got unexpected output:\n  got:    %s\n  expect: %s", got, expect)
	}
}

func TestRenderFormWithDifferentOptions(t *testing.T) {
	renderer := preact10.RenderEngine

	tests := []struct {
		name    string
		data    FormData
		expect  string
	}{
		{
			name: "colors dropdown",
			data: FormData{
				FormTitle:     "Pick a Color",
				InputLabel:    "Item Name",
				InputName:     "item",
				DropdownLabel: "Color",
				DropdownName:  "color",
				DropdownOptions: []DropdownOption{
					{Value: "red", Label: "Red"},
					{Value: "green", Label: "Green"},
					{Value: "blue", Label: "Blue"},
				},
			},
			expect: `<form>` +
				`<h2>Pick a Color</h2>` +
				`<div>` +
				`<label for="item">Item Name</label>` +
				`<input type="text" id="item" name="item" />` +
				`</div>` +
				`<div>` +
				`<label for="color">Color</label>` +
				`<select id="color" name="color">` +
				`<option value="red">Red</option>` +
				`<option value="green">Green</option>` +
				`<option value="blue">Blue</option>` +
				`</select>` +
				`</div>` +
				`<button type="submit">Submit</button>` +
				`</form>`,
		},
		{
			name: "empty dropdown options",
			data: FormData{
				FormTitle:       "Empty Form",
				InputLabel:      "Name",
				InputName:       "name",
				DropdownLabel:   "Options",
				DropdownName:    "opts",
				DropdownOptions: []DropdownOption{},
			},
			expect: `<form>` +
				`<h2>Empty Form</h2>` +
				`<div>` +
				`<label for="name">Name</label>` +
				`<input type="text" id="name" name="name" />` +
				`</div>` +
				`<div>` +
				`<label for="opts">Options</label>` +
				`<select id="opts" name="opts"></select>` +
				`</div>` +
				`<button type="submit">Submit</button>` +
				`</form>`,
		},
		{
			name: "single option",
			data: FormData{
				FormTitle:     "Single Choice",
				InputLabel:    "Description",
				InputName:     "desc",
				DropdownLabel: "Type",
				DropdownName:  "type",
				DropdownOptions: []DropdownOption{
					{Value: "default", Label: "Default Option"},
				},
			},
			expect: `<form>` +
				`<h2>Single Choice</h2>` +
				`<div>` +
				`<label for="desc">Description</label>` +
				`<input type="text" id="desc" name="desc" />` +
				`</div>` +
				`<div>` +
				`<label for="type">Type</label>` +
				`<select id="type" name="type">` +
				`<option value="default">Default Option</option>` +
				`</select>` +
				`</div>` +
				`<button type="submit">Submit</button>` +
				`</form>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := lavish.NewBundle(renderer)
			got, err := bundle.RenderJSX("form.jsx", formJSX, tt.data)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.expect {
				t.Fatalf("got unexpected output:\n  got:    %s\n  expect: %s", got, tt.expect)
			}
		})
	}
}
