package docx

import (
	"github.com/gomutex/godocx/wml/ctypes"
	"github.com/gomutex/godocx/wml/stypes"
)

// DocDefaults holds the document-level defaults from w:docDefaults.
type DocDefaults struct {
	// Run defaults (w:rPrDefault/w:rPr)
	FontASCII    string // w:rFonts @w:ascii
	FontHAnsi    string // w:rFonts @w:hAnsi
	FontEastAsia string // w:rFonts @w:eastAsia
	FontCS       string // w:rFonts @w:cs
	FontSize     int    // w:sz   — half-points (24 = 12 pt)
	FontSizeCS   int    // w:szCs — half-points
	Bold         bool
	Italic       bool
	Strike       bool
	Color        string // hex, e.g. "FF0000" ("auto" → "")
	Language     string // e.g. "en-US"

	// Paragraph defaults (w:pPrDefault/w:pPr)
	SpacingBefore   int // twips
	SpacingAfter    int // twips
	LineSpacing     int // raw OOXML value (240 = single)
	IndentLeft      int // twips
	IndentRight     int // twips
	IndentFirstLine int // twips
	IndentHanging   int // twips
}

// GetDocumentDefaults returns the document-level defaults from w:docDefaults.
func (rd *RootDoc) GetDocumentDefaults() (DocDefaults, error) {
	if rd.DocStyles == nil {
		return DocDefaults{}, nil
	}

	var result DocDefaults

	if dd := rd.DocStyles.DocDefaults; dd != nil {
		if dd.RunProp != nil && dd.RunProp.RunProp != nil {
			rp := dd.RunProp.RunProp
			if rp.Fonts != nil {
				result.FontASCII = rp.Fonts.Ascii
				result.FontHAnsi = rp.Fonts.HAnsi
				result.FontEastAsia = rp.Fonts.EastAsia
				result.FontCS = rp.Fonts.CS
			}
			if rp.Size != nil {
				result.FontSize = int(rp.Size.Value)
			}
			if rp.SizeCs != nil {
				result.FontSizeCS = int(rp.SizeCs.Value)
			}
			result.Bold = onOffToBool(rp.Bold)
			result.Italic = onOffToBool(rp.Italic)
			result.Strike = onOffToBool(rp.Strike)
			if rp.Color != nil && rp.Color.Val != "auto" {
				result.Color = rp.Color.Val
			}
			if rp.Lang != nil && rp.Lang.Val != nil {
				result.Language = *rp.Lang.Val
			}
		}

		if dd.ParaProp != nil && dd.ParaProp.ParaProp != nil {
			pp := dd.ParaProp.ParaProp
			if pp.Spacing != nil {
				if pp.Spacing.Before != nil {
					result.SpacingBefore = int(*pp.Spacing.Before)
				}
				if pp.Spacing.After != nil {
					result.SpacingAfter = int(*pp.Spacing.After)
				}
				if pp.Spacing.Line != nil {
					result.LineSpacing = *pp.Spacing.Line
				}
			}
			if pp.Indent != nil {
				if pp.Indent.Left != nil {
					result.IndentLeft = *pp.Indent.Left
				}
				if pp.Indent.Right != nil {
					result.IndentRight = *pp.Indent.Right
				}
				if pp.Indent.FirstLine != nil {
					result.IndentFirstLine = int(*pp.Indent.FirstLine)
				}
				if pp.Indent.Hanging != nil {
					result.IndentHanging = int(*pp.Indent.Hanging)
				}
			}
		}
	}

	return result, nil
}

// onOffToBool converts a *ctypes.OnOff to bool.
// Per OOXML spec: element present with no val attribute means true.
func onOffToBool(o *ctypes.OnOff) bool {
	if o == nil {
		return false
	}
	if o.Val == nil {
		return true
	}
	switch *o.Val {
	case stypes.OnOffZero, stypes.OnOffFalse, stypes.OnOffOff:
		return false
	default:
		return true
	}
}

// GetStyleByID retrieves a style from the document styles collection based on the given style ID and type.
func (rd *RootDoc) GetStyleByID(styleID string, styleType stypes.StyleType) *ctypes.Style {
	if rd.DocStyles == nil {
		return nil
	}

	for _, style := range rd.DocStyles.StyleList {
		if style.ID == nil || style.Type == nil {
			continue
		}

		if *style.ID == styleID && *style.Type == styleType {
			return &style
		}
	}
	return nil
}

// GetDefaultStyleByType retrieves the default style for a given style type.
func (rd *RootDoc) GetDefaultStyleByType(styleType stypes.StyleType) *ctypes.Style {
	if rd.DocStyles == nil {
		return nil
	}

	for _, style := range rd.DocStyles.StyleList {
		if style.Type == nil || style.Default == nil {
			continue
		}

		if *style.Type == styleType && *style.Default == stypes.OnOff("1") {
			return &style
		}
	}
	return nil
}
