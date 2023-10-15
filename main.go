package main

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Invalid arguments")
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Failed to create directory: %s\n", err)
			}
		}
		headFileContents := []byte("ref: refs/heads/master\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			log.Fatalf("Failed to write file: %s\n", err)
		}
		fmt.Println("Git repository initialized successfully")
	case "cat-file":
		blobSha := os.Args[3]
		catFile(blobSha)
	default:
		log.Fatalf("Unknown command %s\n", command)
	}
}

const objectsPath = ".git/objects"

func catFile(blobSha string) {
	obj := loadObj(blobSha)
	os.Stdout.Write(obj.Data)
}

func loadObj(blobSha string) *Object {
	objPath := filepath.Join(objectsPath, blobSha[:2], blobSha[2:])
	obj, err := os.Open(objPath)
	if err != nil {
		log.Fatalf("Failed to open object file: %s\n", err)
	}
	defer obj.Close()
	zr, err := zlib.NewReader(obj)
	if err != nil {
		panic(err)
	}
	defer zr.Close()
	br := bufio.NewReader(zr)

	// format: (commit|tree|blob) <byte_size>\x00<body>
	kind, err := br.ReadString(' ')
	kind = strings.Trim(kind, " ")
	if err != nil {
		log.Fatalf("Failed to read object: %s\n", err)
	}
	var objType ObjType
	switch kind {
	case "commit":
		objType = Commit
	case "tree":
		objType = Tree
	case "blob":
		objType = Blob
	default:
		log.Fatalf("Unknown object type: %s\n", kind)
	}

	sizes, err := br.ReadString(0)
	sizes = strings.Trim(sizes, "\x00")
	if err != nil {
		panic(err)
	}
	size, err := strconv.ParseInt(sizes, 10, 64)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, size)
	if _, err = io.ReadFull(br, buf); err != nil {
		log.Fatalf("Failed to  read object: %s\n", err)
	}

	return &Object{
		Type: objType,
		Size: size,
		Data: buf,
	}
}

type ObjType int

const (
	Commit ObjType = iota
	Tree
	Blob
)

type Object struct {
	Type ObjType
	Size int64
	Data []byte
}
