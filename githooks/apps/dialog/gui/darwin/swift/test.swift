#!/usr/bin/swift

import Cocoa
import SwiftUI

// MARK: - Constants

let app = NSApplication.shared
var exitCode : Int32 = 1
let origin = CGPoint(
  x: NSScreen.main?.frame.midX ?? 50,
  y: NSScreen.main?.frame.midY ?? 50)

// MARK: - Views

struct DialogButton: View {

  typealias Callback = () -> Void

  private var spacing = 5
  private var ok: Callback
  private var cancel: Callback

  init(ok: @escaping Callback, cancel: @escaping Callback) {
    self.ok = ok
    self.cancel = cancel
  }

  var body: some View {

    // GeometryReader { g in
    HStack(
      spacing: CGFloat(self.spacing)
    ) {

      Button(action: ok) {
        Text("Ok")
          .frame(maxWidth: .infinity, minHeight: 32)
          .foregroundColor(Color.white)
          .background(Color.blue)
          .cornerRadius(5)
      }
      .buttonStyle(PlainButtonStyle())
      Spacer()
      Button(action: cancel) {
        Text("Cancel")
          .frame(maxWidth: .infinity, minHeight: 32)
          .foregroundColor(Color.white)
          .background(Color.gray)
          .cornerRadius(5)
      }
      .buttonStyle(PlainButtonStyle())
    }
    //}
  }

}

let iconDefault =
  "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertNoteIcon.icns"
let errorPath =
  "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns"
let warningPath =
  "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertCautionIcon.icns"
let questionPath =

  "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/GenericQuestionMarkIcon.icns"

func getIcon(_ what: String) -> NSImage? {
  switch what {
  case "info":
    return NSImage(contentsOfFile: iconDefault)
  case "error":
    return NSImage(contentsOfFile: errorPath)
  case "warning":
    return NSImage(contentsOfFile: warningPath) ?? NSImage(named: NSImage.cautionName)
  case "question":
    return NSImage(contentsOfFile: questionPath)
  default:
    return NSImage(contentsOfFile: iconDefault)
  }
}

func getIcon2(_ what: String) -> NSImage? {
  let i = getIcon(what)
  let rep = i!.bestRepresentation(
    for: NSRect(x: 0, y: 0, width: 128, height: 128), context: nil, hints: nil)
  let i2 = NSImage(size: rep!.size)
  i2.addRepresentation(rep!)
  return i2
}

struct OptionView: View {

  struct Option: Identifiable, Hashable {
    let name: String
    var index = 0

    let id = UUID()
  }

  private var options = [Option]()

  init(_ opts: [String]) {
    for (i, s) in opts.enumerated() {
      self.options.append(Option(name: s, index: i))
    }
  }

  @State private var multiSelection = Set<UUID>()

  var body: some View {

    VStack(
      alignment: .leading,
      spacing: 10
    ) {

      HStack(alignment: .center) {
        Spacer()
        Image(nsImage: getIcon2("warning")!).resizable()
          .frame(maxWidth: 64, maxHeight: 64)
          .aspectRatio(contentMode: .fit)
        Spacer()
      }

      VStack(spacing: 5) {
        Text("This is a message").font(.title2).fontWeight(.bold)
        Spacer(minLength: 5)
        Text(
          """
          This is a message laksjd lkajs dlökja södlkj aslökdj alöskjd lakjsd ölkajsd lökjas dlkjasldkj alskdj ölaksjd ölajksd löajks d
          .asjd lkjsad fölkjasöldkfj alsökjdf löasdf
          asdf älkasjd fölkajsdf
          asd flasjdf ölsaj
          """
        ).font(.body).fontWeight(.bold)
      }

      GeometryReader { g in
        ScrollView {
          List(selection: $multiSelection) {
            ForEach(self.options) { o in
              Text("\(o.name)")
            }
          }
          .frame(width: g.size.width, height: g.size.height)
        }
        .border(Color.gray)
      }

      Spacer()
      Text("\(multiSelection.count) selections")

      DialogButton(ok: ok, cancel: cancel)
    }
    .frame(
      minWidth: 0, maxWidth: .infinity, minHeight: 0, maxHeight: .infinity, alignment: .topLeading)
  }

  func ok() {
    let sel = self.options.filter { el in return self.multiSelection.contains(el.id) }
    print(sel.map { "\($0.index)" }.joined(separator: ","))


    exitCode = 0
    app.terminate(self)
  }

  func cancel() {
    exitCode = 1
    app.terminate(self)
  }

}

// MARK: - Setup

class AppDelegate: NSObject, NSApplicationDelegate {

  let window = NSWindow(
    contentRect: NSRect(
      origin: origin,
      size: CGSize(
        width: 300,
        height: 400
      )
    ),
    styleMask: [.resizable, .titled, .closable, .miniaturizable],
    backing: .buffered,
    defer: false,
    screen: nil
  )

  func applicationDidFinishLaunching(_ notification: Notification) {

    let o = OptionView([String](repeating: "Option 1", count: 40))

    // setup the window
    window.titlebarAppearsTransparent = true

    window.makeKeyAndOrderFront(nil)
    NSApp.setActivationPolicy(.regular)
    NSApp.activate(ignoringOtherApps: true)

    window.contentView = NSHostingView(
      rootView: o.padding(20)
    )
  }

  func applicationShouldTerminateAfterLastWindowClosed(
    _ sender: NSApplication
  ) -> Bool {
    return true
  }

  func applicationWillTerminate(_ aNotification: Notification) {
    exit(exitCode)
  }

}

let delegate = AppDelegate()
app.delegate = delegate
app.run()
