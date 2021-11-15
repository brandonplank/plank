//
//  main.swift
//  swiftcombine
//
//  Created by Brandon Plank on 7/20/21.
//

import Foundation
import ArgumentParser
import Swime
import PlankCore

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
    
    @Flag(name: [.customLong("encrypt"), .customShort("e")], help: "Encrypt the data.")
    var encrypt = false
    
    @Option(name: [.customLong("keyiv"), .customShort("k")], help: "Specifies the key and iv to decrypt plank file.")
    var keyiv: String?
    
    @Flag(name: [.customLong("verbose"), .customShort("v")], help: "Show extra logging for debugging purposes")
    var verbose = false
    
    mutating func run() throws {
        var dataToPass = [Data]()
        var ReturnData: PlankCore.Plank.PlankFile?
        var lastData: Data? = nil
        var Filenames = [String]()
        for file in file {
            let fileUrl = URL(fileURLWithPath: file)
            Filenames.append(fileUrl.lastPathComponent)
            do {
                lastData = try Data(contentsOf: fileUrl)
                dataToPass.append(lastData!)
            } catch {
                throw ExitCode.failure
            }
        }
        
        if lastData == nil { throw ExitCode.failure }
        
        if decode {
            if Swime.mimeType(data: lastData!)?.type == .plank {
                var files:[Data]?
                var filenames:[String]?
                var PlankStructure: PlankCore.Plank.PlankFile
                if keyiv != nil {
                    PlankStructure = PlankCore.Plank.Decode().run(lastData!, keyiv: keyiv!)!
                } else {
                    PlankStructure = PlankCore.Plank.Decode().run(lastData!)!
                }
                files = PlankStructure.PlankDataArray
                filenames = PlankStructure.PlankFiles
                
                if files == nil { print("An error happened while decoding"); throw ExitCode.failure }
                if filenames == nil { print("An error happened while reading the file names"); throw ExitCode.failure }
                for (index, file) in files!.enumerated() {
                    do {
                        output = filenames![index]
                        try file.write(to: URL(fileURLWithPath: output!))
                        print("Saved file to: \(output!)")
                    } catch {
                        print("Bad file path!")
                        throw ExitCode.failure
                    }
                }
            } else {
                print("File is not a .plank file!")
                throw ExitCode.failure
            }
        } else {
            if encrypt {
                ReturnData = PlankCore.Plank.Encode().run(dataToPass, filenames: Filenames, encrypt: true)
            } else {
                ReturnData = PlankCore.Plank.Encode().run(dataToPass, filenames: Filenames)
            }
            
            if ReturnData?.PlankData == nil { throw ExitCode.failure }
            
            if output != nil {
                do {
                    output = "\(output!).plank"
                    try ReturnData!.PlankData!.write(to: URL(fileURLWithPath: output!))
                    print("Saved file to: \(output!)")
                } catch {
                    print("Bad file path!")
                    throw ExitCode.failure
                }
                return
            }
        }
    }
}

plank.main()

