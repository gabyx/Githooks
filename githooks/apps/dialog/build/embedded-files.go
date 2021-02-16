// Code generated for package build by go-bindata DO NOT EDIT. (@generated)
// sources:
// osascripts/file.js.tmpl
// osascripts/message.js.tmpl
// osascripts/notify.js.tmpl
// osascripts/options.js.tmpl
// osascripts/entry.js.tmpl
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

var _guiDarwinOsascriptsFileJsTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x8f\xb1\x6a\x03\x31\x0c\x40\xe7\xf3\x57\x98\xd0\x41\x5e\x0c\xfd\x80\x0c\xf7\x05\x1d\x6e\x2c\x1d\x84\xad\xb4\x0a\x87\x2d\x64\x5d\x68\x38\xfc\xef\xc5\xd7\xa6\xed\x64\xf0\x7b\x7a\x42\x37\x54\x8f\x22\xfe\xec\x67\x91\x95\x13\x1a\xd7\x12\xd3\xa6\x4a\xc5\xfe\x7d\x41\x70\x28\x12\xb9\xa4\x75\xcb\xb4\x18\x96\x8c\x9a\xe7\x9c\x79\xd0\xe6\xcf\xde\x74\xa3\xc3\xc1\x64\x7c\x43\x23\x08\xce\x8d\x7c\x15\x1b\x7c\xdf\xaf\xad\x16\x1f\x5f\xc4\x5a\xef\xce\x99\xde\x77\x37\x0d\x41\x69\x70\x14\x79\xfd\x73\x48\x8f\xb5\xbd\xbf\xc1\x98\x0f\x6e\xe2\x8b\x87\x59\x15\xef\x91\xdb\xf1\x82\x52\x0b\xc1\xef\x6e\x9a\x94\x5a\xbc\x56\x2e\xf0\x08\x2c\x24\xa8\x68\x55\x7b\x0f\x6e\xea\x9e\xd6\x46\xbf\xa6\xd5\xc5\x94\xcb\x3b\x0c\xe4\xba\x4f\x68\xe9\xc3\x03\x1d\x2d\xbe\xc0\x0f\xa5\xf0\x38\xb7\xc1\x29\x61\x49\xb4\xae\x94\x4f\x21\x8c\xce\x53\xa4\x4f\x36\x78\xfe\x4e\x7c\x05\x00\x00\xff\xff\xd8\x14\xb3\xaf\x46\x01\x00\x00")

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

	info := bindataFileInfo{name: "gui/darwin/osascripts/file.js.tmpl", size: 326, mode: os.FileMode(420), modTime: time.Unix(1613429800, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _guiDarwinOsascriptsMessageJsTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x8f\xc1\x4a\xc4\x30\x10\x86\xcf\x99\xa7\x08\xc5\x43\x02\x12\xf0\x01\x7a\x58\x3c\x79\x5a\x71\x05\x0f\xe2\x61\x9a\x8c\x36\x4b\x4d\xc2\x64\xba\xae\x94\xbc\xbb\xb4\xec\xaa\xd7\xef\xff\xe7\x9b\x99\xfd\x70\xbc\x77\xf1\xb3\x64\x16\xd3\x55\x09\x53\x1c\x3a\x0b\x70\x42\xd6\x58\x8a\xee\xf5\xae\x94\x29\x7a\x94\x98\x93\xf3\x33\x33\x25\xf9\x87\x8c\x05\x2c\xc5\xc5\xe4\xa7\x39\xd0\x41\x30\x05\xe4\xb0\x0b\x21\xae\x69\xd5\xbd\x16\x9e\x69\xeb\xa0\x97\x78\x42\x21\x73\xd1\xe7\x22\x6b\xbe\x2c\xc7\x9a\x93\x76\xfb\x22\xb5\x35\x58\xa9\xfb\x8a\x32\x3e\xf8\x9c\x74\xaf\x1f\x51\x46\x73\xed\xbc\x5c\x78\x6b\x16\x40\xf8\x7b\x01\xb5\x9a\x98\x56\x11\x96\xf2\xfa\x27\x23\xde\xee\x6b\xed\xed\x77\xfa\x99\xce\xd2\xda\xed\xb6\xd8\x82\x62\xaa\x6e\x98\x45\x72\x7a\x22\x99\x39\x51\x80\xa6\x3d\x8a\x1f\xb5\x21\xab\x17\x50\xf1\xdd\x1c\x84\x63\xfa\x30\x64\xaf\x2f\x56\xd3\x79\x4c\x9e\xa6\x89\x42\x67\xed\x02\x4a\xdd\x38\x3a\x47\x31\x77\x16\x54\x83\xf6\x13\x00\x00\xff\xff\x05\xe6\xc9\xeb\x51\x01\x00\x00")

func guiDarwinOsascriptsMessageJsTmplBytes() ([]byte, error) {
	return bindataRead(
		_guiDarwinOsascriptsMessageJsTmpl,
		"gui/darwin/osascripts/message.js.tmpl",
	)
}

func guiDarwinOsascriptsMessageJsTmpl() (*asset, error) {
	bytes, err := guiDarwinOsascriptsMessageJsTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "gui/darwin/osascripts/message.js.tmpl", size: 337, mode: os.FileMode(420), modTime: time.Unix(1613503807, 0)}
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

var _guiDarwinOsascriptsOptionsJsTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x8e\x3d\x4e\x03\x31\x10\x85\x7b\x9f\x62\x88\x28\xd6\x12\xf2\xc2\x01\x52\xac\x52\x51\xa5\x48\x89\x28\x9c\xf5\x00\x13\x39\xf6\x68\x3c\x8e\x88\x56\xbe\x3b\xf2\x12\x7e\xdc\x58\x7a\xdf\xd3\x37\x6f\x7f\x3c\xed\x1c\x9d\x39\x8b\x0e\x9b\xa2\x21\xd2\x71\x63\x8d\xb9\x78\x01\xcf\x0c\x5b\x98\x98\x23\xcd\x5e\x29\x27\x37\x57\x11\x4c\xfa\x2f\x1a\xac\xf1\xcc\x8e\xd2\x1c\x6b\xc0\x83\xfa\x14\xbc\x84\x29\x04\xea\xb4\xc0\x16\x54\x2a\xae\x1d\x3f\x2b\x5d\xbc\xe2\x70\xd3\x67\xd6\xce\x97\xe5\x54\x72\x02\xb7\x67\x2d\xad\x7d\x23\xc1\x4e\x3c\xf3\xcb\x1f\x45\x59\x0f\xb6\xf6\x3a\xfc\x84\xcf\x8a\xe7\xd2\xda\xc3\xaa\xb2\x86\xde\x60\x98\x44\xfc\xd5\x51\x59\xff\x41\xb0\x58\x0b\x8b\x01\x00\xe8\x54\xb0\xb8\x88\xe9\x5d\x3f\xe0\x6e\x0b\x8f\x16\x16\x18\x47\x98\x8e\x25\xc7\xaa\x18\xaf\x90\x32\xe4\xaa\x5c\xb5\xd7\x23\x15\x05\x2a\x80\x67\xd6\x2b\x08\x6a\x95\x84\x61\x95\xf5\xd7\x65\xa7\x4c\xe9\x77\xce\x01\xd9\x8b\xd7\x2c\xad\xd9\xb5\xd5\x4c\x03\x8c\x05\x6f\x0b\xc6\x11\x76\x3e\xcd\x18\xe3\xcd\x72\xef\xf0\x93\x74\x78\xb2\xa6\x99\xaf\x00\x00\x00\xff\xff\x90\xf3\x1d\xf4\x88\x01\x00\x00")

func guiDarwinOsascriptsOptionsJsTmplBytes() ([]byte, error) {
	return bindataRead(
		_guiDarwinOsascriptsOptionsJsTmpl,
		"gui/darwin/osascripts/options.js.tmpl",
	)
}

func guiDarwinOsascriptsOptionsJsTmpl() (*asset, error) {
	bytes, err := guiDarwinOsascriptsOptionsJsTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "gui/darwin/osascripts/options.js.tmpl", size: 392, mode: os.FileMode(420), modTime: time.Unix(1613510740, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _guiDarwinOsascriptsEntryJsTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x8f\xc1\x4a\x03\x31\x10\x86\xcf\x99\xa7\x08\x8b\x87\x04\x24\xe0\x03\xf4\x50\x3c\x79\xaa\x58\xc1\x83\x78\x98\x26\xa3\x4d\x59\xb3\x61\x32\x5b\x57\x96\x79\x77\xd9\xa5\x55\xaf\xdf\xff\xcf\x37\x33\xbb\xc3\xe9\x3e\xe4\xcf\x3a\xb0\xb8\xae\x49\xea\xf3\xa1\xf3\x00\x67\x64\x8b\xb5\xda\x8d\xdd\xd6\xda\xe7\x88\x92\x87\x12\xe2\xc8\x4c\x45\xfe\x21\xe7\x01\x6b\x0d\xb9\xc4\x7e\x4c\xb4\x17\x2c\x09\x39\x6d\x53\xca\x4b\xda\xec\xc6\x0a\x8f\xb4\x76\x30\x4a\x3e\xa3\x90\xbb\xe8\x87\x2a\x4b\x3e\xcf\xa7\x36\x14\x1b\x76\x55\x9a\x2a\x2c\x34\x7c\x65\x39\x3e\xc4\xa1\xd8\x8d\x7d\x44\x39\xba\x6b\xe7\xe5\xc2\x55\x3d\x80\xf0\xf7\x0c\x66\x31\x31\x2d\x22\xac\xf5\xf5\x4f\x46\xbc\xde\xa7\xfa\xf6\x3b\xfd\x4c\x93\xa8\xde\xae\x8b\x3d\x18\xa6\x16\x84\x26\x79\x22\x19\xb9\x50\x02\xb5\x11\x25\x1e\xad\x23\x6f\x67\x30\xf9\xdd\xed\x85\x73\xf9\x70\xe4\xaf\x0f\x36\xd7\x45\x2c\x91\xfa\x9e\x52\xe7\xfd\x0c\xc6\xdc\x04\x9a\xb2\xb8\x3b\x0f\x46\x41\x7f\x02\x00\x00\xff\xff\x06\x69\x3c\x92\x4f\x01\x00\x00")

func guiDarwinOsascriptsEntryJsTmplBytes() ([]byte, error) {
	return bindataRead(
		_guiDarwinOsascriptsEntryJsTmpl,
		"gui/darwin/osascripts/entry.js.tmpl",
	)
}

func guiDarwinOsascriptsEntryJsTmpl() (*asset, error) {
	bytes, err := guiDarwinOsascriptsEntryJsTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "gui/darwin/osascripts/entry.js.tmpl", size: 335, mode: os.FileMode(420), modTime: time.Unix(1613510753, 0)}
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
	"gui/darwin/osascripts/file.js.tmpl":    guiDarwinOsascriptsFileJsTmpl,
	"gui/darwin/osascripts/message.js.tmpl": guiDarwinOsascriptsMessageJsTmpl,
	"gui/darwin/osascripts/notify.js.tmpl":  guiDarwinOsascriptsNotifyJsTmpl,
	"gui/darwin/osascripts/options.js.tmpl": guiDarwinOsascriptsOptionsJsTmpl,
	"gui/darwin/osascripts/entry.js.tmpl":   guiDarwinOsascriptsEntryJsTmpl,
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
				"entry.js.tmpl":   &bintree{guiDarwinOsascriptsEntryJsTmpl, map[string]*bintree{}},
				"file.js.tmpl":    &bintree{guiDarwinOsascriptsFileJsTmpl, map[string]*bintree{}},
				"message.js.tmpl": &bintree{guiDarwinOsascriptsMessageJsTmpl, map[string]*bintree{}},
				"notify.js.tmpl":  &bintree{guiDarwinOsascriptsNotifyJsTmpl, map[string]*bintree{}},
				"options.js.tmpl": &bintree{guiDarwinOsascriptsOptionsJsTmpl, map[string]*bintree{}},
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
