package docx

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddComment_basic(t *testing.T) {
	rd := setupRootDoc(t)
	para := rd.AddParagraph("Hello world")

	id := rd.AddComment(para, "Great sentence!", "Alice", "A")

	assert.Equal(t, 0, id)
	require.NotNil(t, rd.Comments)
	require.Len(t, rd.Comments.Comments, 1)

	c := rd.Comments.Comments[0]
	assert.Equal(t, 0, c.ID)
	assert.Equal(t, "Alice", c.Author)
	assert.Equal(t, "A", c.Initials)
	assert.NotEmpty(t, c.Date)
	require.Len(t, c.Paragraphs, 1)
	assert.Equal(t, "Great sentence!", c.Paragraphs[0].Children[0].Run.Children[0].Text.Text)
}

func TestAddComment_injectsParagraphMarkers(t *testing.T) {
	rd := setupRootDoc(t)
	para := rd.AddParagraph("Some text")

	id := rd.AddComment(para, "A comment", "Bob", "B")

	ct := para.GetCT()
	// Paragraph should now contain the original run plus 3 appended children
	// (CommentRangeStart, CommentRangeEnd, reference run).
	var starts, ends, refs int
	for _, child := range ct.Children {
		if child.CommentRangeStart != nil && child.CommentRangeStart.ID == id {
			starts++
		}
		if child.CommentRangeEnd != nil && child.CommentRangeEnd.ID == id {
			ends++
		}
		if child.Run != nil {
			for _, rc := range child.Run.Children {
				if rc.CmntRef != nil && rc.CmntRef.ID == id {
					refs++
				}
			}
		}
	}
	assert.Equal(t, 1, starts)
	assert.Equal(t, 1, ends)
	assert.Equal(t, 1, refs)
}

func TestAddComment_multipleIDs(t *testing.T) {
	rd := setupRootDoc(t)
	para := rd.AddParagraph("Text")

	id0 := rd.AddComment(para, "first", "Alice", "A")
	id1 := rd.AddComment(para, "second", "Bob", "B")

	assert.Equal(t, 0, id0)
	assert.Equal(t, 1, id1)
	assert.Len(t, rd.Comments.Comments, 2)
}

func TestRemoveComment(t *testing.T) {
	rd := setupRootDoc(t)
	para := rd.AddParagraph("Some text")

	id := rd.AddComment(para, "to remove", "Alice", "A")

	err := rd.RemoveComment(id)
	require.NoError(t, err)

	assert.Len(t, rd.Comments.Comments, 0)

	ct := para.GetCT()
	for _, child := range ct.Children {
		assert.Nil(t, child.CommentRangeStart)
		assert.Nil(t, child.CommentRangeEnd)
		if child.Run != nil {
			for _, rc := range child.Run.Children {
				assert.Nil(t, rc.CmntRef)
			}
		}
	}
}

func TestRemoveComment_notFound(t *testing.T) {
	rd := setupRootDoc(t)

	err := rd.RemoveComment(999)
	assert.Error(t, err)
}

func TestAddComment_marshalXML(t *testing.T) {
	rd := setupRootDoc(t)
	para := rd.AddParagraph("Hello")
	rd.AddComment(para, "Nice!", "Carol", "C")

	// Marshal the comments to XML
	var buf strings.Builder
	enc := xml.NewEncoder(&buf)
	err := rd.Comments.MarshalXML(enc, xml.StartElement{})
	require.NoError(t, err)
	enc.Flush()

	out := buf.String()
	assert.Contains(t, out, "w:comments")
	assert.Contains(t, out, "w:comment")
	assert.Contains(t, out, `w:author="Carol"`)
	assert.Contains(t, out, `w:initials="C"`)
	assert.Contains(t, out, "Nice!")

	// Marshal the paragraph and verify range markers present
	var paraBuf strings.Builder
	paraEnc := xml.NewEncoder(&paraBuf)
	err = para.GetCT().MarshalXML(paraEnc, xml.StartElement{})
	require.NoError(t, err)
	paraEnc.Flush()

	paraOut := paraBuf.String()
	assert.Contains(t, paraOut, "commentRangeStart")
	assert.Contains(t, paraOut, "commentRangeEnd")
	assert.Contains(t, paraOut, "commentReference")
}

func TestAddComment_ensuresRelationship(t *testing.T) {
	rd := setupRootDoc(t)
	para := rd.AddParagraph("Text")
	rd.AddComment(para, "note", "Dave", "D")

	found := false
	for _, rel := range rd.Document.DocRels.Relationships {
		if rel.Type == "http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments" {
			found = true
			break
		}
	}
	assert.True(t, found, "comments relationship should be registered")

	// Calling AddComment again should not add a second relationship
	rd.AddComment(para, "another", "Eve", "E")
	count := 0
	for _, rel := range rd.Document.DocRels.Relationships {
		if rel.Type == "http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments" {
			count++
		}
	}
	assert.Equal(t, 1, count, "only one comments relationship should exist")
}
