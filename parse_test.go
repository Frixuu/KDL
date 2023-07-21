package kdl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsesSimpleDocument(t *testing.T) {

	input := `
		name "John Smith"
		planet "Earth"
		children {
			daughter "Alice" age=3
			daughter "Laura" --social-media=(lie)false
			son "Michael" {
				likes {
					dinosaurs
					"fire trucks"
				}
			}
		}
	`

	doc, err := ParseString(input)
	assert.NoError(t, err)

	assert.Equal(t, "John Smith", doc.Nodes[0].Args[0].AsString())
	assert.Equal(t, 1, len(doc.Nodes[2].Children[1].Props))
}
