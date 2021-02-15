// Code generated for package build by go-bindata DO NOT EDIT. (@generated)
// sources:
// osascripts/file.js.tmpl
// osascripts/msg.js.tmpl
// osascripts/notify.js.tmpl
package build

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _guiDarwinOsascriptsFileJsTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x8f\x41\x6e\x2b\x21\x0c\x40\xd7\xc3\x29\xac\xe8\x2f\xcc\x06\xe9\x1f\x20\x8b\x39\x43\x96\x55\x17\x16\x38\xad\x23\x04\xc8\x78\xa2\x46\x23\xee\x5e\x31\x6d\xda\xae\x10\xbc\xc7\xb3\x7c\x27\x05\x6a\x0d\xce\xb0\xb6\x96\x25\x92\x49\x2d\x21\x6e\xaa\x5c\xec\xcf\x13\x7a\x47\xad\x05\x29\x31\x6f\x89\x2f\x46\x25\x91\xa6\x35\x25\x99\xb4\xc3\x19\x4c\x37\x3e\x1c\x8a\x26\x77\x32\x46\xef\xdc\xcc\xd7\x66\x93\xef\xfb\xad\xd7\x02\x61\x5e\xc7\x70\xce\xf4\xb1\xbb\x65\x0a\xca\x93\x53\x6b\x2f\xbf\x0e\xeb\x31\x76\x8c\x57\x9c\x1f\xbc\x5b\xe4\x0a\xb8\xaa\xd2\x23\x48\x3f\x4e\x54\xee\xde\xc3\xee\x96\x45\xb9\x87\x5b\x95\x82\xcf\x40\xe7\x46\x4a\x56\x75\x0c\xef\x96\x01\x9c\x3b\xff\x98\x56\x2f\xa6\x52\xde\x70\x22\x37\x20\x92\xc5\x77\x40\x3e\x5a\x72\xc5\x6f\xca\xfe\xb9\x6e\xc7\x53\xa4\x12\x39\x67\x4e\x27\xef\x67\xe7\x5f\xe0\x0f\x31\xfc\xff\x95\xf8\x0c\x00\x00\xff\xff\x46\x08\x5e\xeb\x46\x01\x00\x00")

func guiDarwinOsascriptsFileJsTmplBytes() ([]byte, error) {
	return bindataRead(
		_guiDarwinOsascriptsFileJsTmpl,
		"gui/darwin/osascripts/file.js.tmpl",
	)
}

func guiDarwinOsascriptsFileJsTmpl() (*asset, error) {
	bytes, err := guiDarwinOsascriptsFileJsTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "gui/darwin/osascripts/file.js.tmpl", size: 326, mode: os.FileMode(420), modTime: time.Unix(1613428364, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _guiDarwinOsascriptsMsgJsTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x8f\xc1\x4a\xc4\x30\x10\x86\xcf\x99\xa7\x08\xc5\x43\x02\x12\xf0\x01\x7a\x58\x3c\x79\x5a\x71\x05\x0f\xe2\x61\x9a\x8c\x36\x4b\x4d\xc2\x64\xba\xae\x94\xbc\xbb\xb4\xec\xaa\xd7\xef\xff\xe7\x9b\x99\xfd\x70\xbc\x77\xf1\xb3\x64\x16\xd3\x55\x09\x53\x1c\x3a\x0b\x70\x42\xd6\x58\x8a\xee\xf5\xae\x94\x29\x7a\x94\x98\x93\xf3\x33\x33\x25\xf9\x87\x8c\x05\x2c\xc5\xc5\xe4\xa7\x39\xd0\x41\x30\x05\xe4\xb0\x0b\x21\xae\x69\xd5\xbd\x16\x9e\x69\xeb\xa0\x97\x78\x42\x21\x73\xd1\xe7\x22\x6b\xbe\x2c\xc7\x9a\x93\x76\xfb\x22\xb5\x35\x58\xa9\xfb\x8a\x32\x3e\xf8\x9c\x74\xaf\x1f\x51\x46\x73\xed\xbc\x5c\x78\x6b\x16\x40\xf8\x7b\x01\xb5\x9a\x98\x56\x11\x96\xf2\xfa\x27\x23\xde\xee\x6b\xed\xed\x77\xfa\x99\xce\xd2\xda\xed\xb6\xd8\x82\x62\xaa\x6e\x98\x45\x72\x7a\x22\x99\x39\x51\x80\x06\x1e\xc5\x8f\xda\x90\xd5\x0b\xa8\xf8\x6e\x0e\xc2\x31\x7d\x18\xb2\xd7\x17\xab\xe9\x3c\x26\x4f\xd3\x44\xa1\xb3\x76\x01\xa5\x6e\x1c\x9d\xa3\x98\x3b\x0b\xaa\x41\xfb\x09\x00\x00\xff\xff\x69\xc8\x25\x32\x51\x01\x00\x00")

func guiDarwinOsascriptsMsgJsTmplBytes() ([]byte, error) {
	return bindataRead(
		_guiDarwinOsascriptsMsgJsTmpl,
		"gui/darwin/osascripts/msg.js.tmpl",
	)
}

func guiDarwinOsascriptsMsgJsTmpl() (*asset, error) {
	bytes, err := guiDarwinOsascriptsMsgJsTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "gui/darwin/osascripts/msg.js.tmpl", size: 337, mode: os.FileMode(420), modTime: time.Unix(1613428396, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _guiDarwinOsascriptsNotifyJsTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\xcd\x41\xaa\x02\x31\x0c\xc6\xf1\x7d\x4f\x91\xe5\x1b\x78\xf4\x06\xb3\x98\x0b\xe8\x42\x2f\x10\x9a\x0a\x91\xd2\x84\x34\x1d\x94\xa1\x77\x97\x0a\x03\x6e\xff\x09\xbf\x6f\x47\x03\x54\x85\x15\x36\xd5\xc2\x09\x9d\xa5\xc6\xd4\xcd\x72\xf5\x9f\xf4\xb7\x04\x54\x8d\x5c\x53\xe9\x94\x6f\x8e\x95\xd0\x68\x23\xe2\x79\x6d\xb0\x82\x5b\xcf\x21\x4c\x4f\xd4\x67\x38\x8e\x67\x93\x0a\xf1\xaa\xde\xc6\x08\xbb\x30\xcd\xa9\x48\xdc\xb4\xe0\xfb\x22\xce\x8f\x53\x3f\x7f\xef\xf9\xe5\x63\xfc\x7f\x89\xe5\x13\x00\x00\xff\xff\x1e\x4b\x94\xd0\x9c\x00\x00\x00")

func guiDarwinOsascriptsNotifyJsTmplBytes() ([]byte, error) {
	return bindataRead(
		_guiDarwinOsascriptsNotifyJsTmpl,
		"gui/darwin/osascripts/notify.js.tmpl",
	)
}

func guiDarwinOsascriptsNotifyJsTmpl() (*asset, error) {
	bytes, err := guiDarwinOsascriptsNotifyJsTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "gui/darwin/osascripts/notify.js.tmpl", size: 156, mode: os.FileMode(420), modTime: time.Unix(1613429180, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"gui/darwin/osascripts/file.js.tmpl":   guiDarwinOsascriptsFileJsTmpl,
	"gui/darwin/osascripts/msg.js.tmpl":    guiDarwinOsascriptsMsgJsTmpl,
	"gui/darwin/osascripts/notify.js.tmpl": guiDarwinOsascriptsNotifyJsTmpl,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"gui": &bintree{nil, map[string]*bintree{
		"darwin": &bintree{nil, map[string]*bintree{
			"osascripts": &bintree{nil, map[string]*bintree{
				"file.js.tmpl":   &bintree{guiDarwinOsascriptsFileJsTmpl, map[string]*bintree{}},
				"msg.js.tmpl":    &bintree{guiDarwinOsascriptsMsgJsTmpl, map[string]*bintree{}},
				"notify.js.tmpl": &bintree{guiDarwinOsascriptsNotifyJsTmpl, map[string]*bintree{}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
