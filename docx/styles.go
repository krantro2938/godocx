package docx

import (
	"github.com/gomutex/godocx/wml/ctypes"
	"github.com/gomutex/godocx/wml/stypes"
)

// DocDefaults holds the document-level run defaults from w:docDefaults and
// all named styles defined in the document.
type DocDefaults struct {
	// From w:docDefaults/w:rPrDefault/w:rPr/w:rFonts
	FontFamily         string // @w:ascii
	FontFamilyHAnsi    string // @w:hAnsi
	FontFamilyEastAsia string // @w:eastAsia
	FontFamilyCS       string // @w:cs (complex script)

	// From w:docDefaults/w:rPrDefault/w:rPr/w:sz (half-points; divide by 2 for pt)
	FontSizeHalfPts uint64
	// From w:docDefaults/w:rPrDefault/w:rPr/w:szCs
	FontSizeCSHalfPts uint64

	// All w:style definitions from the document
	Styles []ctypes.Style
}

// GetDocumentDefaults returns the document-level defaults extracted from
// w:docDefaults together with all named styles defined in the document.
func (rd *RootDoc) GetDocumentDefaults() (DocDefaults, error) {
	if rd.DocStyles == nil {
		return DocDefaults{}, nil
	}

	var result DocDefaults

	if dd := rd.DocStyles.DocDefaults; dd != nil {
		if dd.RunProp != nil && dd.RunProp.RunProp != nil {
			rp := dd.RunProp.RunProp
			if rp.Fonts != nil {
				result.FontFamily = rp.Fonts.Ascii
				result.FontFamilyHAnsi = rp.Fonts.HAnsi
				result.FontFamilyEastAsia = rp.Fonts.EastAsia
				result.FontFamilyCS = rp.Fonts.CS
			}
			if rp.Size != nil {
				result.FontSizeHalfPts = rp.Size.Value
			}
			if rp.SizeCs != nil {
				result.FontSizeCSHalfPts = rp.SizeCs.Value
			}
		}
	}

	result.Styles = rd.DocStyles.StyleList

	return result, nil
}

// GetStyleByID retrieves a style from the document styles collection based on the given style ID and type.
//
// Parameters:
//   - styleID: A string representing the ID of the style to retrieve.
//   - styleType: An stypes.StyleType indicating the type of style (e.g., paragraph, character, table).
//
// Returns:
//   - *ctypes.Style: A pointer to the style matching the provided ID and type, if found; otherwise, nil.
//
// This method searches through the document's style list to find a style with the specified ID and type.
// If no matching style is found or if the document styles collection is nil, it returns nil.
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
//
// Parameters:
//   - styleType: An stypes.StyleType indicating the type of style (e.g., paragraph, character, table).
//
// Returns:
//   - *ctypes.Style: A pointer to the default style for the given type, if found; otherwise, nil.
//
// This method searches through the document's style list to find a style marked
// as default (w:default="1") for the specified type.
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
