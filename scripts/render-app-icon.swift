import AppKit

// Package the unchanged artwork inside a padded macOS icon silhouette.
let input = CommandLine.arguments[1]
let output = CommandLine.arguments[2]
guard let source = NSImage(contentsOfFile: input),
      let bitmap = NSBitmapImageRep(bitmapDataPlanes: nil, pixelsWide: 1024, pixelsHigh: 1024,
          bitsPerSample: 8, samplesPerPixel: 4, hasAlpha: true, isPlanar: false,
          colorSpaceName: .deviceRGB, bytesPerRow: 0, bitsPerPixel: 0),
      let context = NSGraphicsContext(bitmapImageRep: bitmap) else {
    fatalError("Cannot load artwork or allocate icon canvas")
}
NSGraphicsContext.saveGraphicsState()
NSGraphicsContext.current = context
context.imageInterpolation = .high
NSColor.clear.setFill()
NSRect(x: 0, y: 0, width: 1024, height: 1024).fill()
let bounds = NSRect(x: 100, y: 100, width: 824, height: 824)
let mask = NSBezierPath(roundedRect: bounds, xRadius: 185, yRadius: 185)
mask.addClip()
source.draw(in: bounds, from: .zero, operation: .copy, fraction: 1)
context.flushGraphics()
NSGraphicsContext.restoreGraphicsState()
guard let png = bitmap.representation(using: .png, properties: [:]) else {
    fatalError("Cannot encode icon")
}
try png.write(to: URL(fileURLWithPath: output))
