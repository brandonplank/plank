// swift-tools-version:5.5
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "plank",
    platforms: [
            .macOS(.v10_12)
    ],
    dependencies: [
        .package(url: "https://github.com/apple/swift-argument-parser", from: "0.4.0"),
        .package(url: "https://github.com/krzyzanowskim/CryptoSwift.git", .upToNextMajor(from: "1.4.1")),
        .package(url: "https://github.com/sendyhalim/Swime", .branch("master")),
        .package(url: "https://github.com/brandonplank/PlankCore", .branch("main"))
        
    ],
    targets: [
        // Targets are the basic building blocks of a package. A target can define a module or a test suite.
        // Targets can depend on other targets in this package, and on products in packages this package depends on.
        .executableTarget(
            name: "plank",
            dependencies: ["PlankCore", "Swime", "CryptoSwift", .product(name: "ArgumentParser", package: "swift-argument-parser")]),
        .testTarget(
            name: "plankTests",
            dependencies: ["plank"]),
    ]
)
