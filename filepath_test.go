package commons

import "testing"

func TestStemExtWithSize(t *testing.T) {
	tests := []struct {
		name       string
		maxExtSize int
		wantStem   string
		wantExt    string
	}{
		{
			name:       "test/hello.doc",
			maxExtSize: 3,
			wantStem:   "hello",
			wantExt:    ".doc",
		},
		{
			name:       "test/hello.docx",
			maxExtSize: 3,
			wantStem:   "hello.docx",
			wantExt:    "",
		},
		{
			name:       "test/hello.docx",
			maxExtSize: 4,
			wantStem:   "hello",
			wantExt:    ".docx",
		},
		{
			name:       "test/hello.doc.gz",
			maxExtSize: 3,
			wantStem:   "hello",
			wantExt:    ".doc.gz",
		},
		{
			name:       "test/hello.docx.gz",
			maxExtSize: 3,
			wantStem:   "hello.docx",
			wantExt:    ".gz",
		},
		{
			name:       "test/hello.docx.gz",
			maxExtSize: 4,
			wantStem:   "hello",
			wantExt:    ".docx.gz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStem, gotExt := StemExtWithSize(tt.name, tt.maxExtSize)
			if gotStem != tt.wantStem {
				t.Errorf("StemExtWithSize() gotStem = %v, want %v", gotStem, tt.wantStem)
			}
			if gotExt != tt.wantExt {
				t.Errorf("StemExtWithSize() gotExt = %v, want %v", gotExt, tt.wantExt)
			}
		})
	}
}
