package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"github.com/akamensky/argparse"
	"github.com/brandonplank/PlankCore"
)

func main() {
	parser := argparse.NewParser("plank", "Brandon Planks' custom archive filetype written in GO.")

	var verbose *bool = parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "Prints more info"})
	var output *string = parser.String("o", "output", &argparse.Options{Required: false, Help: "Send to .plank file"})
	var extract *bool = parser.Flag("e", "extract", &argparse.Options{Required: false, Help: "Extracts the plank file"})
	var files *[]os.File = parser.FileList("f", "files", os.O_RDWR, 0600, &argparse.Options{Required: true, Help: "Files to be passed to the program"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	if *verbose {
		fmt.Println("Running in verbose")
	}

	type data []byte

	var filenames 	[]string
	var readFiles	[]data

	if files != nil {
		for index, item := range *files {
			file, err := item.Stat()
			if err != nil {
				panic(err)
			}
			data, err := ioutil.ReadAll(&item)
			if err != nil {
				panic(err)
			}
			defer item.Close()

			if *verbose {
				fmt.Printf("File: %d\tItem: %s\tSize: 0x%x\n", index+1, item.Name(), file.Size())
			}

			filenames = append(filenames, item.Name())
			readFiles = append(readFiles, data)
		}
	} else {
		panic("Error with files")
	}

	if *output != "" {
		data := plankcore.PlankEncode(readFiles, filenames, *verbose)
			
		if *verbose {
			fmt.Printf("Encoded\n")
			fmt.Printf("%s", hex.Dump(data))
		}

		fmt.Printf("Writing to %s\n", *output)

		err := os.WriteFile(*output, data, 0644)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Wrote %s\n", *output)
	}

	if *extract {
		out := plankcore.PlankDecode(readFiles[0], *verbose)
		if *verbose {
			fmt.Printf("Decoded\n")
		}
		
		for i := 0; i < len(out.Data); i++ {
			filename := out.Filenames[i]
			data     := out.Data[i]
			fmt.Printf("Writing to %s\n", filename)

			err := os.WriteFile(filename, data, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}
