package kdl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const inputSimple string = `
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

func TestParsesSimpleDocument(t *testing.T) {
	doc, err := ParseString(inputSimple)
	assert.NoError(t, err)
	assert.Equal(t, "John Smith", doc.Nodes[0].Args[0].StringValue())
	lauraProps := doc.Nodes[2].Children[1].Props
	assert.Equal(t, 1, len(lauraProps))
	assert.Equal(t, false, lauraProps["--social-media"].BoolValue())
}

func BenchmarkParseSimpleDocument(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParseString(inputSimple)
	}
}
