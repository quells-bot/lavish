package lavish_test

import (
	"strings"
	"testing"

	"github.com/quells/lavish"
	"github.com/quells/lavish/engine/preact10"
	"github.com/stretchr/testify/require"
)

// TestSharedComponentInBundle demonstrates loading a reusable component
// into a Bundle and using it across multiple templates.
func TestSharedComponentInBundle(t *testing.T) {
	// Define a shared Card component that can be reused in multiple templates
	cardComponent := `
		const Card = ({ title, content, variant = "default" }) => (
			<div class={"card card-" + variant}>
				<h3 class="card-title">{title}</h3>
				<p class="card-content">{content}</p>
			</div>
		);
	`

	// Create a bundle with the Card component loaded as a shared module
	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(
		renderer,
		lavish.ComponentJSX("card.jsx", cardComponent, renderer.GetJSXOptions()),
	)

	t.Run("single card", func(t *testing.T) {
		template := `
			const { title, content } = data;
			render(<Card title={title} content={content} />);
		`

		data := map[string]string{
			"title":   "Welcome",
			"content": "This is a card component",
		}

		got, err := bundle.RenderJSX("single-card.jsx", template, data)
		require.NoError(t, err)

		// Verify the output contains expected elements
		require.Contains(t, got, `<div class="card card-default">`)
		require.Contains(t, got, `<h3 class="card-title">Welcome</h3>`)
		require.Contains(t, got, `<p class="card-content">This is a card component</p>`)
	})

	t.Run("multiple cards in list", func(t *testing.T) {
		template := `
			const { cards } = data;
			render(
				<div class="card-list">
					{cards.map(card => (
						<Card
							title={card.title}
							content={card.content}
							variant={card.variant}
						/>
					))}
				</div>
			);
		`

		data := map[string]interface{}{
			"cards": []map[string]string{
				{"title": "First Card", "content": "Content 1", "variant": "primary"},
				{"title": "Second Card", "content": "Content 2", "variant": "secondary"},
				{"title": "Third Card", "content": "Content 3", "variant": "default"},
			},
		}

		got, err := bundle.RenderJSX("card-list.jsx", template, data)
		require.NoError(t, err)

		// Verify all cards are rendered
		require.Contains(t, got, `<div class="card-list">`)
		require.Contains(t, got, `card-primary`)
		require.Contains(t, got, `card-secondary`)
		require.Contains(t, got, `card-default`)
		require.Contains(t, got, "First Card")
		require.Contains(t, got, "Second Card")
		require.Contains(t, got, "Third Card")
	})

	t.Run("nested layout with shared component", func(t *testing.T) {
		// Load an additional layout component
		layoutComponent := `
			const Layout = ({ children }) => (
				<html lang="en">
					<head>
						<title>Card Gallery</title>
					</head>
					<body>
						<div class="container">
							{children}
						</div>
					</body>
				</html>
			);
		`

		bundleWithLayout := lavish.NewBundle(
			renderer,
			lavish.ComponentJSX("card.jsx", cardComponent, renderer.GetJSXOptions()),
			lavish.ComponentJSX("layout.jsx", layoutComponent, renderer.GetJSXOptions()),
		)

		template := `
			const { title, description } = data;
			render(
				<Layout>
					<h1>Gallery</h1>
					<Card title={title} content={description} variant="featured" />
				</Layout>
			);
		`

		data := map[string]string{
			"title":       "Featured Item",
			"description": "This card is nested in a layout",
		}

		got, err := bundleWithLayout.RenderJSX("page.jsx", template, data)
		require.NoError(t, err)

		// Verify the complete structure
		require.Contains(t, got, `<html lang="en">`)
		require.Contains(t, got, `<title>Card Gallery</title>`)
		require.Contains(t, got, `<div class="container">`)
		require.Contains(t, got, `<h1>Gallery</h1>`)
		require.Contains(t, got, `card-featured`)
		require.Contains(t, got, "Featured Item")
		require.Contains(t, got, "This card is nested in a layout")
	})
}

// TestSharedComponentWithFragments demonstrates using Fragment
// with shared components.
func TestSharedComponentWithFragments(t *testing.T) {
	// Define components that use Fragment
	components := `
		const ListItem = ({ text, index }) => (
			<li>
				<strong>{index}.</strong> {text}
			</li>
		);

		const Section = ({ title, items }) => (
			<Fragment>
				<h2>{title}</h2>
				<ul>
					{items.map((item, idx) => (
						<ListItem text={item} index={idx + 1} />
					))}
				</ul>
			</Fragment>
		);
	`

	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(
		renderer,
		lavish.ComponentJSX("components.jsx", components, renderer.GetJSXOptions()),
	)

	template := `
		const { sections } = data;
		render(
			<div class="document">
				{sections.map(section => (
					<Section title={section.title} items={section.items} />
				))}
			</div>
		);
	`

	data := map[string]interface{}{
		"sections": []map[string]interface{}{
			{
				"title": "Features",
				"items": []string{"Fast", "Simple", "Powerful"},
			},
			{
				"title": "Benefits",
				"items": []string{"Easy to use", "Well tested", "Great performance"},
			},
		},
	}

	got, err := bundle.RenderJSX("document.jsx", template, data)
	require.NoError(t, err)

	// Verify sections are rendered (Fragment shouldn't add extra wrappers)
	require.Contains(t, got, `<h2>Features</h2>`)
	require.Contains(t, got, `<h2>Benefits</h2>`)
	require.Contains(t, got, `<strong>1.</strong> Fast`)
	require.Contains(t, got, `<strong>3.</strong> Powerful`)

	// Verify no extra Fragment wrappers
	require.NotContains(t, strings.ToLower(got), "fragment")
}

// TestMultipleSharedComponents demonstrates loading multiple
// independent shared components.
func TestMultipleSharedComponents(t *testing.T) {
	buttonComponent := `
		const Button = ({ label, onClick, variant = "primary" }) => (
			<button class={"btn btn-" + variant} onclick={onClick}>
				{label}
			</button>
		);
	`

	inputComponent := `
		const Input = ({ name, placeholder, value = "" }) => (
			<input
				type="text"
				name={name}
				placeholder={placeholder}
				value={value}
				class="form-input"
			/>
		);
	`

	formComponent := `
		const Form = ({ action, method = "POST", children }) => (
			<form action={action} method={method} class="form">
				{children}
			</form>
		);
	`

	renderer := preact10.RenderEngine
	bundle := lavish.NewBundle(
		renderer,
		lavish.ComponentJSX("button.jsx", buttonComponent, renderer.GetJSXOptions()),
		lavish.ComponentJSX("input.jsx", inputComponent, renderer.GetJSXOptions()),
		lavish.ComponentJSX("form.jsx", formComponent, renderer.GetJSXOptions()),
	)

	template := `
		const { formAction, fields } = data;
		render(
			<Form action={formAction}>
				<div class="form-group">
					<Input
						name="username"
						placeholder="Enter username"
						value={fields.username}
					/>
				</div>
				<div class="form-group">
					<Input
						name="email"
						placeholder="Enter email"
						value={fields.email}
					/>
				</div>
				<div class="form-actions">
					<Button label="Submit" variant="primary" />
					<Button label="Cancel" variant="secondary" />
				</div>
			</Form>
		);
	`

	data := map[string]interface{}{
		"formAction": "/submit",
		"fields": map[string]string{
			"username": "john_doe",
			"email":    "john@example.com",
		},
	}

	got, err := bundle.RenderJSX("contact-form.jsx", template, data)
	require.NoError(t, err)

	// Verify form structure
	require.Contains(t, got, `<form action="/submit" method="POST" class="form">`)

	// Verify inputs
	require.Contains(t, got, `name="username"`)
	require.Contains(t, got, `value="john_doe"`)
	require.Contains(t, got, `name="email"`)
	require.Contains(t, got, `value="john@example.com"`)

	// Verify buttons
	require.Contains(t, got, `class="btn btn-primary"`)
	require.Contains(t, got, `class="btn btn-secondary"`)
	require.Contains(t, got, ">Submit</button>")
	require.Contains(t, got, ">Cancel</button>")
}
