package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type scaledTheme struct {
	base fyne.Theme
}

func (t *scaledTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x19, G: 0x1a, B: 0x17, A: 0xff}
	case theme.ColorNameHeaderBackground, theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x20, G: 0x22, B: 0x1e, A: 0xff}
	case theme.ColorNameButton, theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x2a, G: 0x2d, B: 0x28, A: 0xff}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x22, G: 0x25, B: 0x20, A: 0xff}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0x4a, G: 0x51, B: 0x47, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xe8, G: 0xe6, B: 0xde, A: 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x9a, G: 0x9d, B: 0x91, A: 0xff}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0xb6, G: 0xb4, B: 0xa9, A: 0xff}
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return color.NRGBA{R: 0x9a, G: 0xbf, B: 0x7f, A: 0xff}
	case theme.ColorNameForegroundOnPrimary, theme.ColorNameForegroundOnSuccess, theme.ColorNameForegroundOnWarning:
		return color.NRGBA{R: 0x11, G: 0x13, B: 0x0f, A: 0xff}
	case theme.ColorNameForegroundOnError:
		return color.NRGBA{R: 0xff, G: 0xfb, B: 0xf7, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x9a, G: 0xbf, B: 0x7f, A: 0x4d}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xf2, A: 0x12}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0x9a, G: 0xbf, B: 0x7f, A: 0x55}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x9a, G: 0xbf, B: 0x7f, A: 0x42}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0x3b, G: 0x3f, B: 0x38, A: 0xff}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x66}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0xc4, G: 0xc8, B: 0xb8, A: 0xb8}
	case theme.ColorNameScrollBarBackground:
		return color.NRGBA{R: 0x23, G: 0x25, B: 0x20, A: 0xff}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x8d, G: 0xbd, B: 0x86, A: 0xff}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xd8, G: 0xb3, B: 0x5a, A: 0xff}
	case theme.ColorNameError:
		return color.NRGBA{R: 0xd3, G: 0x6d, B: 0x6d, A: 0xff}
	default:
		return t.base.Color(name, variant)
	}
}

func (t *scaledTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *scaledTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *scaledTheme) Size(name fyne.ThemeSizeName) float32 {
	base := t.base.Size(name)
	// Scale up sizes for readability
	switch name {
	case theme.SizeNameText:
		return base * 1.5 // ~21px instead of ~14px
	case theme.SizeNameHeadingText:
		return base * 1.3
	case theme.SizeNameInputRadius:
		return base * 1.2
	case theme.SizeNameInlineIcon:
		return base * 1.3
	case theme.SizeNamePadding:
		return base * 1.5
	case theme.SizeNameScrollBar:
		return base * 1.3
	case theme.SizeNameScrollBarSmall:
		return base * 1.3
	case theme.SizeNameSeparatorThickness:
		return base * 1.5
	default:
		return base * 1.2
	}
}

// NewTheme returns a theme with scaled-up sizes for readability.
func NewTheme() fyne.Theme {
	return &scaledTheme{base: theme.DarkTheme()}
}
