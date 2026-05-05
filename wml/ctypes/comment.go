package ctypes

import (
	"encoding/xml"
	"strconv"
)

// Comment represents a single review comment (CT_Comment) stored in word/comments.xml.
type Comment struct {
	ID         int
	Author     string
	Initials   string
	Date       string // ISO 8601 timestamp
	Paragraphs []*Paragraph
}

func (c Comment) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "w:comment"
	start.Attr = append(start.Attr,
		xml.Attr{Name: xml.Name{Local: "w:id"}, Value: strconv.Itoa(c.ID)},
		xml.Attr{Name: xml.Name{Local: "w:author"}, Value: c.Author},
	)
	if c.Initials != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "w:initials"}, Value: c.Initials})
	}
	if c.Date != "" {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "w:date"}, Value: c.Date})
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, p := range c.Paragraphs {
		if err := p.MarshalXML(e, xml.StartElement{}); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

func (c *Comment) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			id, err := strconv.Atoi(attr.Value)
			if err != nil {
				return err
			}
			c.ID = id
		case "author":
			c.Author = attr.Value
		case "initials":
			c.Initials = attr.Value
		case "date":
			c.Date = attr.Value
		}
	}

loop:
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch elem := tok.(type) {
		case xml.StartElement:
			switch elem.Name.Local {
			case "p":
				p := &Paragraph{}
				if err = d.DecodeElement(p, &elem); err != nil {
					return err
				}
				c.Paragraphs = append(c.Paragraphs, p)
			default:
				if err = d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			break loop
		}
	}
	return nil
}

// Comments is the root element of word/comments.xml.
type Comments struct {
	Comments []Comment
}

var commentsNSAttrs = []xml.Attr{
	{Name: xml.Name{Local: "xmlns:w"}, Value: "http://schemas.openxmlformats.org/wordprocessingml/2006/main"},
}

func (cs Comments) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "w:comments"
	start.Attr = append(start.Attr, commentsNSAttrs...)

	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, c := range cs.Comments {
		if err := c.MarshalXML(e, xml.StartElement{}); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

func (cs *Comments) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
loop:
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch elem := tok.(type) {
		case xml.StartElement:
			switch elem.Name.Local {
			case "comment":
				c := Comment{}
				if err = d.DecodeElement(&c, &elem); err != nil {
					return err
				}
				cs.Comments = append(cs.Comments, c)
			default:
				if err = d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			break loop
		}
	}
	return nil
}
