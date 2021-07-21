//
//  plank.swift
//  swiftcombine
//
//  Created by Brandon Plank on 7/21/21.
//

import Foundation

extension Array where Element == UInt8 {
    func asciiRep() -> String {
        var ret: String = ""
        for i in 0..<self.count {
            ret.append(String(format: "%c", self[i]))
        }
        return ret
    }
}

extension Data {
    func conglobberate(_ data: Data...) -> [Data] {
        var ret = [Data]()
        for data in data {
            ret.append(data)
        }
        return ret
    }
}

extension Array {
    func split() -> (left: [Element], right: [Element]) {
        let ct = self.count
        let half = ct / 2
        let leftSplit = self[0 ..< half]
        let rightSplit = self[half ..< ct]
        return (left: Array(leftSplit), right: Array(rightSplit))
    }
    func chunked(into size: Int) -> [[Element]] {
        return stride(from: 0, to: count, by: size).map {
            Array(self[$0 ..< Swift.min($0 + size, count)])
        }
    }
}

extension Int {
    func addrByteArray() -> [UInt8] {
        let data: Data? = withUnsafeBytes(of: self) { Data($0) }
        let byte = [UInt8](data!)
        return byte
    }
}

class Plank {
    struct StaticValues {
        let AddresserSize: Int = 16 // 16 Bytes
    }
    class Encode {
        init() {
        }
        /*
         Defines
         */
        var DataStartOffset: Int = 0
        var DataSections: Int = 0
        var DataSectionSizes = [Int]()
        var DataCombined = [UInt8]()
        var FormedData = [UInt8]()
        var StartOffsets = [Int]()
        var EndOffsets = [Int]()
        var SectionSizes = [Int]()
        
        var One = MemoryLayout<Int>.size // 8 Bytes
        var Two: Int = 0
        

        func run(_ data: [Data]) -> Data {
            // Init the amount of data sections
            DataSections = data.count
            Two = DataSections * Plank.StaticValues().AddresserSize
            for data in data {
                DataSectionSizes.append(data.count)
                let convertedData = [UInt8](data)
                DataCombined.append(contentsOf: convertedData)
            }
            
            /*
            print(DataSectionSizes)
            print(DataCombined)
             */
            
            SectionSizes.append(One)
            SectionSizes.append(Two)
            
            // Static offsets
            StartOffsets.append(0x00) // Start of section 1
            EndOffsets.append(0x07) // End of section 1
            StartOffsets.append(0x08) // Start of section 2
            
            // Dynamic offsets
            EndOffsets.append(One + Two - 1) // End of section 2
            
            // Fill in Section sizes
            
            for i in 0..<DataSections {
                SectionSizes.append(DataSectionSizes[i])
            }
            
            // Fill in Start offsets
            for i in 0..<SectionSizes.count {
                if i > 1 {
                    let addArray = Array(SectionSizes[0...i])
                    EndOffsets.append(addArray.reduce(0, +)-1)
                }
            }
            
            // Fill in Start offsets
            for i in 0..<SectionSizes.count {
                if i > 1 {
                    let addArray = Array(SectionSizes[0...i])
                    StartOffsets.append(addArray.reduce(0, +) - DataSectionSizes[i - 2])
                }
            }
            /*
            for sizes in SectionSizes {
                print(String(format: "size 0x%02x", sizes))
            }
            print("")
            for offset in StartOffsets {
                print(String(format: "start 0x%02x", offset))
            }
            print("")
            for offset in EndOffsets {
                print(String(format: "end 0x%02x", offset))
            }
            */
            
            let SectionOne = withUnsafeBytes(of: Two) { Data($0) }
            FormedData.append(contentsOf: SectionOne)

            // Construct section 2
            for i in 0..<DataSections {
                let startaddr = StartOffsets[i + 2]
                FormedData.append(contentsOf: startaddr.addrByteArray())
                let endaddr = EndOffsets[i + 2]
                FormedData.append(contentsOf: endaddr.addrByteArray())
            }
            
            FormedData.append(contentsOf: DataCombined)
            return Data(FormedData)
        }
    }
    
    class Decode {
        init() {
        }
        /*
         Defines
         */
        var DataStartOffset: Int = 0
        var DataSections: Int = 0
        var DataSectionSizes = [Int]()
        var DataCombined = [UInt8]()
        var FormedData = [UInt8]()
        var CombinedOffsets = [Int]()
        var StartOffsets = [Int]()
        var EndOffsets = [Int]()
        var SectionSizes = [Int]()
        
        var ReturnData = [Data]()
        
        var One = MemoryLayout<Int>.size // 8 Bytes
        var Two: Int = 0
        
        func run(_ data: Data) -> [Data] {
            let SectionOne = data[0...One]
            Two = SectionOne.withUnsafeBytes {
                $0.load(as: Int.self).littleEndian
            }
            print("Section 2 size is: \(Two) bytes")
            let SectionTwo = Array(data[8...Two+One-1])
            //print(SectionTwo)
            DataSections = Two / Plank.StaticValues().AddresserSize
            
            let SectionTwoRead = SectionTwo.chunked(into: 8)
            //print(SectionTwoRead)
            
            for i in 0..<SectionTwoRead.count {
                var trackArray = [UInt8]()
                for j in 0..<SectionTwoRead[i].count {
                    trackArray.append(contentsOf: [UInt8](arrayLiteral: SectionTwoRead[i][j]))
                    //print("\(String(format: "0x%02x", SectionTwoRead[i][j]))")
                    if j == SectionTwoRead[i].count - 1 {
                        let dataoffset = trackArray.withUnsafeBytes {
                            $0.load(as: Int.self).littleEndian
                        }
                        //print(String(format: "0x%0llx", dataoffset))
                        CombinedOffsets.append(dataoffset)
                        trackArray = [UInt8]()
                    }
                }
            }
            
            //print(CombinedOffsets)
            
            // Seperate offsets
            for (index, item) in CombinedOffsets.enumerated() {
                if index.isMultiple(of: 2) {
                    StartOffsets.append(Int(item))
                } else {
                    EndOffsets.append(Int(item))
                }
            }
            #if DEBUG
            print("Start offsets: \(StartOffsets.map { String(format: "0x%02x", $0) }.joined(separator: " "))")
            print("End offsets: \(EndOffsets.map { String(format: "0x%02x", $0) }.joined(separator: " "))")
            #endif
            
            for i in 0..<DataSections {
                ReturnData.append(data[StartOffsets[i]...EndOffsets[i]])
            }
            return ReturnData
        }
    }
}
