package pinfo

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
)

// Consts for Magic numbers and Machine Codes
const (
	BIT64 = 0x8664
	BIT32 = 0x14c
	PE32  = 0x10B
	PE32P = 0x20B
)

// BasicProps contains basic data about a PE file
type BasicProps struct {
	Name   string
	MD5    string
	SHA1   string
	SHA256 string
	//Imphash      string
	//SSDEEP       string
	FileType  string
	Magic     string
	FSize     string
	Libraries []string
	Symbols   []string
	Sections  []pe.Section
	ModelRes  bool
}

// NewProps returns a pointer to a basicProps struct
func NewProps(file *os.File, useModel bool) *BasicProps {
	props := BasicProps{}
	props.Name = file.Name()
	props.fillHashes(file)
	props.fillFileType(file)
	props.fillMagic(file)
	props.fillFileSize(file)
	props.fillLibraries(file)
	props.fillSymbols(file)
	props.fillSections(file)
	if useModel {
		props.fillFromModel(file)
	}

	return &props
}

func (p *BasicProps) fillHashes(file *os.File) {
	mh := md5.New()
	_, err := io.Copy(mh, file)
	if err != nil {
		log.Fatalln("Error copying file into md5 hash:", err)
	}
	mhbytes := mh.Sum(nil)
	p.MD5 = hex.EncodeToString(mhbytes[:])

	s1h := sha1.New()
	file.Seek(0, 0)
	_, err = io.Copy(s1h, file)
	if err != nil {
		log.Fatalln("error copying file into sha1 hash:", err)
	}
	s1hbytes := s1h.Sum(nil)
	p.SHA1 = hex.EncodeToString(s1hbytes)

	s2h := sha256.New()
	file.Seek(0, 0)
	_, err = io.Copy(s2h, file)
	if err != nil {
		log.Fatalln("Error copying file into sha256 hash:", err)
	}
	s2hbytes := s2h.Sum(nil)
	p.SHA256 = hex.EncodeToString(s2hbytes)

	file.Seek(0, 0)
}

func (p *BasicProps) fillFileType(file *os.File) {
	exe, err := pe.NewFile(file)
	if err != nil {
		log.Fatalln("Error converting to pe:", err)
	}
	if exe.Machine == BIT32 {
		p.FileType = "Win32 Exe"
	} else if exe.Machine == BIT64 {
		p.FileType = "Win64 Exe"
	} else {

	}
}

func (p *BasicProps) fillMagic(file *os.File) {
	exe, err := pe.NewFile(file)
	if err != nil {
		log.Fatalln("Error converting to pe:", err)
	}

	magic := exe.OptionalHeader.(*pe.OptionalHeader32).Magic
	if magic == PE32 {
		p.Magic = "PE32"
	} else if magic == PE32P {
		p.Magic = "PE32P"
	} else {
		p.Magic = "Unknown"
	}
}

func (p *BasicProps) fillFileSize(f *os.File) {
	info, err := f.Stat()
	if err != nil {
		log.Fatalf("Error calling stat on file: %s \n", err)
	}
	p.FSize = strconv.FormatInt(info.Size(), 10)
}

func (p *BasicProps) fillSymbols(f *os.File) {
	exe, err := pe.NewFile(f)
	if err != nil {
		fmt.Println("Error converting to PE in fillSymbols")
	}

	p.Symbols, err = exe.ImportedSymbols()
	if err != nil {
		fmt.Println("Error getting imported symbols")
	}
}

func (p *BasicProps) fillLibraries(f *os.File) {
	exe, err := pe.NewFile(f)
	if err != nil {
		fmt.Println("Error converting to PE in fillImports")
	}

	p.Libraries, err = exe.ImportedLibraries()
	if err != nil {
		fmt.Println("Error getting imported libraries")
	}
}

func (p *BasicProps) fillSections(f *os.File) {
	exe, err := pe.NewFile(f)
	if err != nil {
		fmt.Println("Error converting to pe in fillSections")
	}
	for _, val := range exe.Sections {
		p.Sections = append(p.Sections, *val)
	}
}

func (p *BasicProps) fillFromModel(f *os.File) {
	pmodel := exec.Command("python3", "prediction.py", f.Name())
	out, err := pmodel.Output()
	if err != nil {
		fmt.Println("Error running model")
	}
	switch res, _ := strconv.Atoi(string(out)); res {
	case 0:
		p.ModelRes = false
	case 1:
		p.ModelRes = true
	case -1:
		fmt.Println("Error from the model")
	}
}

func (p *BasicProps) String() string {
	return fmt.Sprintf("---Basic Info---\n%-15s%s\n%-15s%s\n%-15s%s\n%-15s%s\n%-15s%s\n%-15s%s", "MD5 Hash: ", p.MD5, "SHA1 Hash: ", p.SHA1, "SHA256 Hash: ", p.SHA256, "File Type: ", p.FileType, "Magic: ", p.Magic, "File Size: ", p.FSize)
}

func (p *BasicProps) ExportHTML() error {
	t, err := template.ParseFiles("binpage.html")
	if err != nil {
		return err
	}

	t.ExecuteTemplate(os.Stdout, "binpage.html", *p)
	return nil
}
