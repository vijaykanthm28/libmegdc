package ubuntu

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/megamsys/urknall/utils"
)

// The "FileCommand" is used to write files to the host being provisioned. The go templating mechanism (see
// http://golang.org/pkg/text/template) is applied on the file's content using the package. Thereby it is possible to
// have dynamic content (based on the package's configuration) for the file content and at the same time store it in
// an asset (which is generated at compile time). Please note that the underlying actions will panic if either no path
// or content are given.
type FileCommand struct {
	Path        string      // Path to the file to create.
	Content     string      // Content of the file to create.
	Owner       string      // Owner of the file to create (root per default).
	Permissions os.FileMode // Permissions of the file created (only changed from system default if set).
}

func (cmd *FileCommand) Render(i interface{}) {
	cmd.Path = utils.MustRenderTemplate(cmd.Path, i)
	cmd.Content = utils.MustRenderTemplate(cmd.Content, i)
}

func (cmd *FileCommand) Validate() error {
	if cmd.Path == "" {
		return fmt.Errorf("no path given")
	}

	if cmd.Content == "" {
		return fmt.Errorf("no content given for file %q", cmd.Path)
	}

	return nil
}

// Helper method to create a file at the given path with the given content, and with owner and permissions set
// accordingly. The "Owner" and "Permissions" options are optional in the sense that they are ignored if set to go's
// default value.
func WriteFile(path string, content string, owner string, permissions os.FileMode) *FileCommand {
	return &FileCommand{Path: path, Content: content, Owner: owner, Permissions: permissions}
}

var b64 = base64.StdEncoding

func (fc *FileCommand) Shell() string {
	buf := &bytes.Buffer{}

	// Zip the content.
	zipper := gzip.NewWriter(buf)
	zipper.Write([]byte(fc.Content))
	zipper.Flush()
	zipper.Close()

	// Encode the zipped content in Base64.
	encoded := b64.EncodeToString(buf.Bytes())

	// Compute sha256 hash of the encoded and zipped content.
	hash := sha256.New()
	hash.Write([]byte(fc.Content))

	// Create temporary filename (hash as filename).
	tmpPath := fmt.Sprintf("/tmp/wunderscale.%x", hash.Sum(nil))

	// Get directory part of target file.
	dir := filepath.Dir(fc.Path)

	// Create command, that will decode and unzip the content and write to the temporary file.
	cmd := ""
	cmd += fmt.Sprintf("mkdir -p %s", dir)
	cmd += fmt.Sprintf(" && echo %s | base64 -d | gunzip > %s", encoded, tmpPath)
	if fc.Owner != "" { // If owner given, change accordingly.
		cmd += fmt.Sprintf(" && chown %s %s", fc.Owner, tmpPath)
	}
	if fc.Permissions > 0 { // If mode given, change accordingly.
		cmd += fmt.Sprintf(" && chmod %o %s", fc.Permissions, tmpPath)
	}
	// Move the temporary file to the requested location.
	cmd += fmt.Sprintf(" && mv %s %s", tmpPath, fc.Path)
	return cmd
}

func (fc *FileCommand) Logging() string {
	sList := []string{"[FILE   ]"}

	if fc.Owner != "" && fc.Owner != "root" {
		sList = append(sList, fmt.Sprintf("[CHOWN:%s]", fc.Owner))
	}

	if fc.Permissions != 0 {
		sList = append(sList, fmt.Sprintf("[CHMOD:%.4o]", fc.Permissions))
	}

	sList = append(sList, " "+fc.Path)

	cLen := len(fc.Content)
	if cLen > 50 {
		cLen = 50
	}
	//sList = append(sList, fmt.Sprintf(" << %s", strings.Replace(string(fc.Content[0:cLen]), "\n", "⁋", -1)))
	return strings.Join(sList, "")
}

type FileSendCommand struct {
	Source      string
	Target      string
	Owner       string
	Permissions os.FileMode
}

func SendFile(source, target, owner string, perm os.FileMode) *FileSendCommand {
	return &FileSendCommand{
		Source:      source,
		Target:      target,
		Owner:       owner,
		Permissions: perm,
	}
}

func (fsc *FileSendCommand) Render(i interface{}) {
	fsc.Source = utils.MustRenderTemplate(fsc.Source, i)
	fsc.Target = utils.MustRenderTemplate(fsc.Target, i)
}

func (fsc *FileSendCommand) Validate() error {
	if fsc.Source == "" {
		return fmt.Errorf("no source path given")
	}

	if _, e := os.Stat(fsc.Source); e != nil {
		return e
	}

	if fsc.Target == "" {
		return fmt.Errorf("no target path given for file %q", fsc.Source)
	}

	return nil
}

func (fsc *FileSendCommand) sourceHash() string {
	fh, e := os.Open(fsc.Source)
	if e != nil {
		panic(e)
	}
	defer fh.Close()

	hash := sha1.New()
	if _, e = io.Copy(hash, fh); e != nil {
		panic(e)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func (fsc *FileSendCommand) Shell() string {
	sList := []string{
		fmt.Sprintf("echo %q", fsc.sourceHash()), // nope use content hash
		fmt.Sprintf("cat - > %s", fsc.Target),
	}

	if fsc.Owner != "root" {
		sList = append(sList, fmt.Sprintf("chown %s %s", fsc.Owner, fsc.Target))
	}
	sList = append(sList, fmt.Sprintf("chmod %s %s", fsc.Permissions, fsc.Target))
	return strings.Join(sList, " && ")
}

func (fsc *FileSendCommand) Input() io.ReadCloser {
	fh, e := os.Open(fsc.Source)
	if e != nil {
		panic(e)
	}
	return fh
}

func (fsc *FileSendCommand) Logging() string {
	sList := []string{"[FILE   ]"}

	if fsc.Owner != "" && fsc.Owner != "root" {
		sList = append(sList, fmt.Sprintf("[CHOWN:%s]", fsc.Owner))
	}

	if fsc.Permissions != 0 {
		sList = append(sList, fmt.Sprintf("[CHMOD:%.4o]", fsc.Permissions))
	}

	sList = append(sList, fmt.Sprintf(" Writing local file %s to %s", fsc.Source, fsc.Target))

	return strings.Join(sList, "")
}
