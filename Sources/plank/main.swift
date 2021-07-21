//
//  main.swift
//  swiftcombine
//
//  Created by Brandon Plank on 7/20/21.
//

import Foundation
import ArgumentParser
import Swime

// Ok here are my notes, going to keep it
/*
 Section one
 Byte 1 - 8 will contain the size of the size mapping section
 Will hold a 64 bit number
 | 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 |
 
 Example
 | 0x39 0x39 0x39 0x39 0x39 0x39 0x39 0x39 | Int
 
 
 Section two, size will be based off the number in section one
 Needs to be in 16 byte intervals as that is how big a section addresser is
 
 Section 2 will contain the amount of sections in section 3-max
 | 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Beginning of sec3 address
 | 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | End of sec3 address
 
 | 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Beginning of sec4 address
 | 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | End of sec4 address
 
 etc
 
 Section 3
 Some data
 | 0x70 0x6f 0x67 0x67 0x65 0x72 0x73 0x00 | p o g g e r s
 
 Section 4
 Some data
 | 0x73 0x77 0x69 0x66 0x74 0x73 0x77 0x69 | s w i f t s w i
 
 end of sizeof(sec1) + sizeof(sec2) is Beginning of sec3 address(the beginning of sec3)
 
 sizeof(sec1) + sizeof(sec2) + sizeof(sec3) is the end of sec3
 
 so the current setup would look like
 | 0x20 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | 32 bytes in section 2, as defined here take this and devide by 16 and we have 2 data sections
 | 0x29 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 3 start address
 | 0x30 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 3 end address
 | 0x31 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 4 start address
 | 0x38 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 4 end address
 | 0x70 0x6f 0x67 0x67 0x65 0x72 0x73 0x00 | Data 1
 | 0x73 0x77 0x69 0x66 0x74 0x73 0x77 0x69 | Data 2
 
 Note data size can differ along with section numbers
 
 When decoding it, we first look at the first 8 bytes
 | 0x20 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | <- This shows us how many bytes is in section 2
 
 From here, we cal calculate how many files/data are stored in our filetype
 There are 16 bytes in each section definition, the start and end addresses 8 bytes for the start, 8 bytes for the end
 
 So, if section 2 has 32 bytes, we know there are 2 files in the data section
 Formula
 
 files = bytes / 16
 
 Now, all we have to do is read all the values in
 | 0x29 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 3 start address
 | 0x30 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 3 end address
 | 0x31 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 4 start address
 | 0x38 0x00 0x00 0x00 0x00 0x00 0x00 0x00 | Section 4 end address
 
 They are 8 bytes each (64 bit)
 
 Once we have that, we have the offsets for the actual files data sections
 
 | 0x70 0x6f 0x67 0x67 0x65 0x72 0x73 0x00 | Data 1
 | 0x73 0x77 0x69 0x66 0x74 0x73 0x77 0x69 | Data 2
 
 These can have different data, i am just doing a simple demo, here is a more complex one with different lens
 
 | 0x54 0x68 0x69 0x73 0x20 0x69 0x73 0x20 | This is
 | 0x70 0x72 0x65 0x74 0x74 0x79 0x20 0x69 | pretty i
 | 0x6e 0x74 0x65 0x72 0x65 0x73 0x74 0x69 | nteresti
 | 0x6e 0x67 0x2c 0x20 0x69 0x74 0x73 0x20 | ng, its
 | 0x73 0x6f 0x20 0x63 0x6f 0x6f 0x6c 0x20 | so cool
 | 0x77 0x68 0x61 0x74 0x20 0x63 0x6f 0x64 | what cod
 | 0x65 0x20 0x63 0x61 0x6e 0x20 0x64 0x6f | e can do
 | 0x53 0x77 0x69 0x66 0x74 0x20 0x69 0x73 | Swift is
 | 0x20 0x67 0x6f 0x6f 0x64 0x2c 0x20 0x66 |  good, f
 | 0x69 0x72 0x65 0x20 0x69 0x73 0x20 0x62 | ire is b
 | 0x61 0x64 0x20 0x6c 0x6d 0x61 0x6f 0x00 | ad lmao
 
 And it has no problems distinguising between any data!
 
 Update
 
 we fully automated, nice
 
 Command line tool time
 */

/*
let string1 = "Testing your mom out lmao".data(using: .utf8)
let string2 = "Lmao".data(using: .utf8)
let string3 = "REALLY LONG REALLY LONG REALLY LONG".data(using: .utf8)
let string4 = "JAPANESEこんにちは世界こんにちは世界こんにちは世界こんにちは世界".data(using: .utf8)

print("Encoding all of this data")
let testing = Plank.Encode().run(Data().conglobberate(string1!, string2!, string3!, string4!))
print([UInt8](testing).asciiRep())
print("Decoding all of this data")
let testing2 = Plank.Decode().run(testing)
print(testing2)
for lol in testing2 {
    print(String(data: lol, encoding: .utf8)!)
}
 */

struct plank: ParsableCommand {
    static let configuration = CommandConfiguration(
        abstract: "A Swift command-line tool to for a .plank file."
    )
    
    @Option(name: [.customLong("file"), .customShort("f")], help: "Specifies the files.")
    var file: [String] = []
    
    @Option(name: [.customLong("output"), .customShort("o")], help: "Specifies the output file. (.plank)")
    var output: String?
    
    @Flag(name: [.customLong("decode"), .customShort("d")], help: "Decode the file. (.plank)")
    var decode = false
    
    @Flag(name: [.customLong("verbose"), .customShort("v")], help: "Show extra logging for debugging purposes")
    var verbose = false
    
    mutating func run() throws {
        var dataToPass = [Data]()
        var ReturnData: Data? = nil
        var lastData: Data? = nil
        for file in file {
            let fileUrl = URL(fileURLWithPath: file)
            //print(fileUrl)
            do {
                lastData = try Data(contentsOf: fileUrl)
                dataToPass.append(lastData!)
            } catch {
                throw ExitCode.failure
            }
        }
        
        
        
        if(decode){
            let files = Plank.Decode().run(lastData!)
            
            var index: Int = 0
            for file in files {
                do {
                    let mimeType = Swime.mimeType(bytes: [UInt8](file))
                    let fileExt = mimeType?.ext
                    if fileExt == nil {
                        output = "\(index)"
                    } else {
                        output = "\(index).\(fileExt ?? "txt")"
                    }
                    try file.write(to: URL(fileURLWithPath: output!))
                    print("Saved file to: \(output!)")
                } catch {
                    print("Bad file path!")
                    throw ExitCode.failure
                }
                index+=1
            }
            
        } else {
            ReturnData = Plank.Encode().run(dataToPass)
            
            if ReturnData == nil { throw ExitCode.failure }
            
            if((output) != nil){
                if output!.contains(".plank"){
                    do {
                        try ReturnData!.write(to: URL(fileURLWithPath: output!))
                        print("Saved file to: \(output!)")
                    } catch {
                        print("Bad file path!")
                        throw ExitCode.failure
                    }
                } else {
                    print("Must save as a .swift file!")
                    throw ExitCode.failure
                }
                return
            }
        }
    }
}

plank.main()

