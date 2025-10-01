package commons

import "path/filepath"

// StemExt is a variant of filepath.Ext that allows extended extension to be detected while also returning the stem.
//
// For example, `filepath.Ext("/path/to/file.tar.gz")` would return ".gz", but `xy3.StemAndExt("/path/to/file.tar.gz")`
// would return ".tar.gz" for the extension, "file" for the stem. This is useful when passed to OpenExclFile:
// "file-1.tar.gz" is more natural than "file.tar-1.gz".
//
// StemExt will only accept file extensions of 5 characters or fewer, so if there is no `.` in the last 6 characters,
// the returned ext will be empty string unlike filepath.Ext which will keep searching until the last path separator or
// `.` is found. That means longer extensions such as ".jfif-tbnl" or ".turbot" will not be found by StemExt but can
// be found by filepath.Ext. Use StemExtWithSize if you need to customise the extension's size.
func StemExt(path string) (stem, ext string) {
	return StemExtWithSize(path, 6)
}

// StemExtWithSize is a variant of StemExt that allows customisation of the extension's size.
//
// If maxExtSize is 3, ".doc" may be returned but ".docx" will not. Similarly, ".doc.gz" may be returned, but if path
// ends in ".docx.gz", only ".gz" is returned.
func StemExtWithSize(path string, maxExtSize int) (stem, ext string) {
	n := len(path) - 1
	for i, j := n, max(0, n-maxExtSize); i >= j; i-- {
		switch path[i] {
		case '\\', '/':
			stem = path[i+1:]
			return
		case '.':
			ext = path[i:] + ext
			path = path[:i]
			n = len(path)
			i, j = n, max(0, n-maxExtSize-1)
			continue
		}
	}

	stem = filepath.Base(path)
	return
}
