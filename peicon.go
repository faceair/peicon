package peicon

import (
	"debug/pe"
	"encoding/binary"
	"io"
	"io/ioutil"
)

type File struct {
	p *pe.File
}

func Open(name string) (*File, error) {
	p, err := pe.Open(name)
	return &File{p}, err
}

func New(r io.ReaderAt) (*File, error) {
	p, err := pe.NewFile(r)
	return &File{p}, err
}

type imageResourceDirectory struct {
	Characteristics      uint32
	TimeDateStamp        uint32
	MajorVersion         uint16
	MinorVersion         uint16
	NumberOfNamedEntries uint16
	NumberOfIdEntries    uint16
}

type imageResourceDirectoryEntry struct {
	NameId uint32
	Data   uint32
}

type imageResourceDataEntry struct {
	Data     uint32
	Size     uint32
	CodePage uint32
	Reserved uint32
}

type ResourceType uint32

// https://msdn.microsoft.com/fr-fr/library/windows/desktop/ms648009(v=vs.85).aspx
const (
	ResourceTypeNone ResourceType = 0
	ResourceTypeIcon ResourceType = 3
)

func (f *File) Icon() ([]byte, error) {
	sect := f.p.Section(".rsrc")

	var iconData []byte
	var iconSize uint32

	var readDirectory func(offset uint32, level int, resourceType ResourceType) error
	readDirectory = func(offset uint32, level int, resourceType ResourceType) error {

		br := io.NewSectionReader(sect, int64(offset), int64(sect.Size)-int64(offset))
		ird := new(imageResourceDirectory)
		err := binary.Read(br, binary.LittleEndian, ird)
		if err != nil {
			return err
		}

		for i := uint16(0); i < ird.NumberOfNamedEntries+ird.NumberOfIdEntries; i++ {
			irde := new(imageResourceDirectoryEntry)
			err = binary.Read(br, binary.LittleEndian, irde)
			if err != nil {
				return err
			}

			if irde.NameId&0x80000000 > 0 {
				continue
			}

			id := irde.NameId & 0xffff

			if irde.Data&0x80000000 > 0 {
				offset := irde.Data & 0x7fffffff
				recResourceType := resourceType
				if level == 0 {
					recResourceType = ResourceType(id)
				}

				err := readDirectory(offset, level+1, recResourceType)
				if err != nil {
					return err
				}
				continue
			}

			dbr := io.NewSectionReader(sect, int64(irde.Data), int64(sect.Size)-int64(irde.Data))

			irda := new(imageResourceDataEntry)
			err = binary.Read(dbr, binary.LittleEndian, irda)
			if err != nil {
				return err
			}

			if resourceType == ResourceTypeIcon {
				dataStart := int64(irda.Data - sect.VirtualAddress)
				sr := io.NewSectionReader(sect, dataStart, int64(irda.Size))

				data, err := ioutil.ReadAll(sr)
				if err != nil {
					return err
				}
				if irda.Size > iconSize {
					iconData = data
				}
			}
		}
		return nil
	}

	err := readDirectory(0, 0, ResourceTypeNone)
	return iconData, err
}
