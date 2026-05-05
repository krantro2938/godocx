package docx

import (
	"fmt"
	"time"

	"github.com/gomutex/godocx/common/constants"
	"github.com/gomutex/godocx/wml/ctypes"
)

// ensureComments initializes rd.Comments and registers the comments.xml relationship
// the first time a comment is added.
func (rd *RootDoc) ensureComments() {
	if rd.Comments != nil {
		return
	}
	rd.Comments = &ctypes.Comments{}
	rd.Document.addRelation(constants.SourceRelationshipComments, "comments.xml")
}

// nextCommentID returns an ID one higher than the current maximum comment ID.
func (rd *RootDoc) nextCommentID() int {
	max := -1
	for _, c := range rd.Comments.Comments {
		if c.ID > max {
			max = c.ID
		}
	}
	return max + 1
}

// AddComment anchors a review comment at the end of para.
// The comment body is the plain text supplied in text.
// Returns the assigned comment ID, which can be passed to RemoveComment.
func (rd *RootDoc) AddComment(para *Paragraph, text, author, initials string) int {
	rd.ensureComments()
	id := rd.nextCommentID()

	comment := ctypes.Comment{
		ID:         id,
		Author:     author,
		Initials:   initials,
		Date:       time.Now().UTC().Format(time.RFC3339),
		Paragraphs: []*ctypes.Paragraph{ctypes.AddParagraph(text)},
	}
	rd.Comments.Comments = append(rd.Comments.Comments, comment)

	refRun := &ctypes.Run{
		Property: &ctypes.RunProperty{
			Style: &ctypes.CTString{Val: "CommentReference"},
		},
		Children: []ctypes.RunChild{
			{CmntRef: &ctypes.Markup{ID: id}},
		},
	}

	para.ct.Children = append(para.ct.Children,
		ctypes.ParagraphChild{CommentRangeStart: &ctypes.Markup{ID: id}},
		ctypes.ParagraphChild{CommentRangeEnd: &ctypes.Markup{ID: id}},
		ctypes.ParagraphChild{Run: refRun},
	)

	return id
}

// RemoveComment deletes the comment with the given ID along with its range markers
// and reference run from every paragraph in the document body.
func (rd *RootDoc) RemoveComment(id int) error {
	if rd.Comments == nil {
		return fmt.Errorf("comment %d not found", id)
	}

	found := false
	kept := rd.Comments.Comments[:0]
	for _, c := range rd.Comments.Comments {
		if c.ID == id {
			found = true
		} else {
			kept = append(kept, c)
		}
	}
	if !found {
		return fmt.Errorf("comment %d not found", id)
	}
	rd.Comments.Comments = kept

	for i := range rd.Document.Body.Children {
		if rd.Document.Body.Children[i].Para != nil {
			removeCommentMarkersFromPara(&rd.Document.Body.Children[i].Para.ct, id)
		}
	}
	return nil
}

func removeCommentMarkersFromPara(para *ctypes.Paragraph, id int) {
	kept := para.Children[:0]
	for _, child := range para.Children {
		switch {
		case child.CommentRangeStart != nil && child.CommentRangeStart.ID == id:
			// drop
		case child.CommentRangeEnd != nil && child.CommentRangeEnd.ID == id:
			// drop
		case child.Run != nil && runHasCommentRef(child.Run, id):
			// drop
		default:
			kept = append(kept, child)
		}
	}
	para.Children = kept
}

func runHasCommentRef(run *ctypes.Run, id int) bool {
	for _, child := range run.Children {
		if child.CmntRef != nil && child.CmntRef.ID == id {
			return true
		}
	}
	return false
}
