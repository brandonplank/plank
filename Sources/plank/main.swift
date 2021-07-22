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
    
    @Flag(name: [.customLong("verbose"), .customShort("v")], help: "Show extra logging for debugging purposes")
    var verbose = false
    
    mutating func run() throws {
        var dataToPass = [Data]()
        var ReturnData: Data? = nil
        var lastData: Data? = nil
        for file in file {
            let fileUrl = URL(fileURLWithPath: file)
            do {
                lastData = try Data(contentsOf: fileUrl)
                dataToPass.append(lastData!)
            } catch {
                throw ExitCode.failure
            }
        }
        
        if lastData == nil { throw ExitCode.failure }
        
        if(decode){
            if(Swime.mimeType(data: lastData!)?.type == .plank){
                let files = PlankCore.Plank.Decode().run(lastData!)
                var index: Int = 0
                for file in files {
                    do {
                        let mimeType = Swime.mimeType(bytes: [UInt8](file))
                        let fileExt = mimeType?.ext
                        if fileExt == nil {
                            output = "\(index)"
                        } else {
                            output = "\(index).\(fileExt!)"
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
                print("File is not a .plank file!")
                throw ExitCode.failure
            }
        } else {
            ReturnData = PlankCore.Plank.Encode().run(dataToPass)
            
            if ReturnData == nil { throw ExitCode.failure }
            
            if((output) != nil){
                do {
                    output = "\(output!).plank"
                    try ReturnData!.write(to: URL(fileURLWithPath: output!))
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

