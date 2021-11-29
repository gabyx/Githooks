#!/usr/bin/env swift

// This script can be called by
// swift githooks/apps/dialog/gui/darwin/swift/message.swift error warning "title" "message" ok cancel 3 A B C

import AppKit
import Foundation

let app = NSApplication.shared
app.setActivationPolicy(.regular)  // Magic to accept keyboard input and be docked!

let args = CommandLine.arguments

let iconDefault =
  "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertNoteIcon.icns"
let errorPath =
  "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns"
let warningPath =
  "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertCautionIcon.icns"
let questionPath =
  "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/GenericQuestionMarkIcon.icns"

final class StandardErrorOutputStream: TextOutputStream {
  func write(_ string: String) {
    FileHandle.standardError.write(Data(string.utf8))
  }
}

var errorStream = StandardErrorOutputStream()

func die<T>(_ arg: T...) {
  print(arg.map({ "\($0)" }).joined(separator: "\n"), to: &errorStream)
  exit(1)
}

func debug<T>(_ arg: T...) {
  print(arg.map({ "\($0)" }).joined(separator: "\n"), to: &errorStream)
}

func getIcon(what: String) -> NSImage? {
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

struct DialogError: Error, LocalizedError {
  let errorDescription: String?

  init(_ description: String) {
    errorDescription = description
  }
}

struct Settings {
  var appIcon: String
  var icon: String
  var title: String
  var text: String
  var okButton: String
  var cancelButton: String

  var width: Int?
  var height: Int?
}

class ExtraButtons {
  var buttons = [String]()
}

class ListOptions {
  var options = [String]()
  var defaultOption = 0
  var multiple = false
}

typealias ParseFunc = ([String]) throws -> Int
typealias ExtraParser = [String: ParseFunc]

func parseArguments(_ args: [String], extra: ExtraParser?) throws -> Settings {
  if args.count < 5 {
    throw DialogError("Wrong arguments given (alert settings).")
  }

  var settings = Settings(
    appIcon: args[0],
    icon: args[1],
    title: args[2],
    text: args[3],
    okButton: args[4],
    cancelButton: args[5])

  var ex = ExtraParser()
  extra?.forEach { (key, value) in ex[key] = value }

  ex["--width"] = { (args: [String]) -> Int in
    settings.width = Int(args[0]) ?? nil
    return 1
  }
  ex["--height"] = { (args: [String]) -> Int in
    settings.height = Int(args[0]) ?? nil
    return 1
  }

  var idx = 6
  while idx < args.count {
    let keyword = args[idx]
    let fun = ex[keyword]
    if fun == nil {
      throw DialogError("Argument '\(args[idx])' not handled.")
    }
    let consumed = try fun!([String](args[(idx + 1)...]))
    idx += consumed + 1
  }

  return settings
}

func parseArray(_ args: [String]) throws -> [String] {
  if args.count == 0 {
    return [String]()
  }

  let nCount = Int(args[0]) ?? 0
  if args.count < nCount + 1 {
    throw DialogError("parseArray: Wrong argument count: : \(args.count)")
  }
  return [String](args[1..<1 + nCount])
}

func parseListOptions(_ extra: ListOptions) throws -> ParseFunc {
  return { (args: [String]) -> Int in
    extra.options = try parseArray(args)
    if extra.options.count + 1 > args.count {
      throw DialogError("parseListOptions: Wrong argument count: \(args.count).")
    }
    extra.defaultOption = Int(args[extra.options.count]) ?? 0
    extra.multiple = args[extra.options.count] == "true" ? true : false
    return (extra.options.count + 1) + 2
  }
}

func parseExtraButtons(_ extra: ExtraButtons) throws -> ParseFunc {
  return { (args: [String]) -> Int in
    extra.buttons = try parseArray(args)
    return extra.buttons.count + 1
  }
}

func setDockIcon(_ settings: Settings) {

  // Set dock icon
  app.dockTile.contentView = NSImageView(image: getIcon(what: settings.appIcon)!)
  app.dockTile.display()
}

func makeDefaultAlert(_ settings: Settings, extraButtons: ExtraButtons? = nil) -> NSAlert {
  let a = NSAlert()
  a.messageText = settings.title
  a.icon = getIcon(what: settings.icon)!
  a.alertStyle = NSAlert.Style.informational
  a.addButton(withTitle: settings.okButton)
  a.addButton(withTitle: settings.cancelButton)

  a.buttons[0].keyEquivalent = "\r"
  a.buttons[1].keyEquivalent = "\u{1b}"

  if extraButtons != nil {
    for b in extraButtons!.buttons {
      a.addButton(withTitle: b)
    }
  }

//   // Text as accessory view since
//   // Height can not be changed
//   let v = NSStackView(frame: NSRect(x: 0, y: 0, width: settings.width ?? 300, height: 500))
//   //v.distribution = NSStackView.Distribution.fill
//   v.autoresizingMask = [.width, .height]
//   let text = NSText(frame: NSRect(x: 0, y: 0, width: settings.width ?? 150, height: 40))
//   text.autoresizingMask = [.width, .height]
//   text.isEditable = false
//   text.string = settings.text
//   text.drawsBackground = false
//   v.addSubview(text)

//   let text2 = NSText(frame: NSRect(x: 0, y: 0, width: settings.width ?? 150, height: 40))
//   text.autoresizingMask = [.width, .height]
//   text2.isEditable = false
//   text2.string = settings.text
//   text2.drawsBackground = false
//   v.addSubview(text2)

//   a.accessoryView = v

  return a
}

func runMessageDialog(args: [String]) throws -> Int {

  let extraButtons = ExtraButtons()
  let settings = try parseArguments(
    args, extra: ["--extra-buttons": parseExtraButtons(extraButtons)])

  setDockIcon(settings)
  let a = makeDefaultAlert(settings, extraButtons: extraButtons)
  app.activate(ignoringOtherApps: true)

  var r = a.runModal().rawValue

  r -= 1000
  if r >= 0 && r <= extraButtons.buttons.count + 2 - 1 {
    print(r)
    return r == 1 ? r : 0  // Cancel -> return 1
  }

  throw DialogError("Wrong button returned.")
}

class TableViewController: NSViewController, NSTableViewDelegate, NSTableViewDataSource {

  var initialized = false
  var frame = NSRect()
  var opts = ListOptions()
  var defaultOption = 0
  var tableView: NSTableView? = nil

  convenience init(frame: NSRect, options: ListOptions) {
    self.init()
    self.frame = frame
    self.opts = options
    self.tableView = NSTableView(frame: frame)
    self.view = self.tableView!
  }

  override func viewDidLayout() {
    if !initialized {
      initialized = true
      setupTableView()
    }
  }

  func setupTableView() {
    self.tableView!.delegate = self
    self.tableView!.dataSource = self
    self.tableView!.headerView = nil
    self.tableView!.allowsMultipleSelection = self.opts.multiple
    self.tableView!.usesAutomaticRowHeights = true
    self.tableView!.backgroundColor = NSColor.clear
    self.tableView!.selectionHighlightStyle = NSTableView.SelectionHighlightStyle.regular
    self.tableView!.selectRowIndexes(
      IndexSet(integer: self.opts.defaultOption), byExtendingSelection: false)
    //tableView!.appearance = NSAppearance(named: NSAppearance.Name.vibrantDark)

    let col = NSTableColumn(identifier: NSUserInterfaceItemIdentifier(rawValue: "col"))
    col.minWidth = 200
    tableView!.addTableColumn(col)
  }

  func numberOfRows(in tableView: NSTableView) -> Int {
    return self.opts.options.count
  }

  func tableView(_ tableView: NSTableView, viewFor tableColumn: NSTableColumn?, row: Int)
    -> NSView?
  {
    let text = NSTextField()
    text.stringValue = self.opts.options[row]
    text.isEditable = false
    text.drawsBackground = false
    text.isBordered = false
    let cell = NSTableCellView()
    cell.addSubview(text)

    text.translatesAutoresizingMaskIntoConstraints = false
    cell.addConstraint(
      NSLayoutConstraint(
        item: text, attribute: .centerY, relatedBy: .equal, toItem: cell,
        attribute: .centerY,
        multiplier: 1, constant: 0))
    cell.addConstraint(
      NSLayoutConstraint(
        item: text, attribute: .left, relatedBy: .equal, toItem: cell, attribute: .left,
        multiplier: 1, constant: 0))
    cell.addConstraint(
      NSLayoutConstraint(
        item: text, attribute: .right, relatedBy: .equal, toItem: cell, attribute: .right,
        multiplier: 1, constant: 0))

    return cell
  }

  func tableView(_ tableView: NSTableView, rowViewForRow row: Int) -> NSTableRowView? {
    let rowView = NSTableRowView()
    rowView.isEmphasized = true
    return rowView
  }
}

func runOptionsDialog(args: [String]) throws -> Int {

  let listOpts = ListOptions()
  let settings = try parseArguments(args, extra: ["--options": parseListOptions(listOpts)])

  setDockIcon(settings)
  let a = makeDefaultAlert(settings)

  let scrollView = NSScrollView(
    frame: NSRect(x: 0, y: 0, width: settings.width ?? 400, height: 200))
  scrollView.hasVerticalScroller = true
  scrollView.hasHorizontalScroller = true
  let clipView = NSClipView(frame: scrollView.bounds)
  clipView.autoresizingMask = [.width, .height]

  let con = TableViewController(frame: clipView.bounds, options: listOpts)
  let tableView = con.tableView
  tableView!.autoresizingMask = [.width, .height]
  clipView.documentView = tableView
  scrollView.contentView = clipView
  a.accessoryView = scrollView

  app.activate(ignoringOtherApps: true)
  var r = a.runModal().rawValue

  r -= 1000
  if r == 0 {
    print(con.tableView!.selectedRowIndexes.map({ "\($0)" }).joined(separator: ","))
    return 0
  } else if r == 1 {
    return 1
  }

  throw DialogError("Wrong button returned.")
}

func run() throws -> Int {
  if args.count == 1 {
    throw DialogError("No arguments given.")
  }

  let type = args[1]
  switch type {
  case "message":
    return try runMessageDialog(args: [String](args[2...]))
  case "options":
    return try runOptionsDialog(args: [String](args[2...]))
  case "demo":

    let text =
      "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet."

    try runOptionsDialog(args: [
      "info", "warning", "title", text, "ok", "cancel", "--options", "3", "Option 1",
      "Option 2",
      "Option 3", "1",
      "false",
    ])

    // try runMessageDialog(args: [
    //   "info", "question", "title", text, "Ok", "Cancel",
    //   "--width", "250",
    // ])

    // try runMessageDialog(args: [
    //   "info", "question", "title", text, "Ok", "Cancel", "--extra-buttons", "3", "Option 1",
    //   "Option 2",
    //   "Option 3",
    //   "--width", "450",
    // ])

    return 0

  default:
    throw DialogError("Cannot show dialog type '\(type)'.")
  }
}

do {
  let exitCode = try run()
  exit(Int32(exitCode))
} catch let e as DialogError {
  die(
    "Dialog error: \(e.errorDescription ?? "nil")."
  )
} catch {
  die("Runtime error: \(error):\n\(Thread.callStackSymbols.forEach{print($0)})")
}
