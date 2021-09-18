#!/usr/bin/env swift

import AppKit
import Foundation

let app = NSApplication.shared
app.setActivationPolicy(.regular)  // Magic to accept keyboard input and be docked!

struct DialogError: Error, LocalizedError {
  let errorDescription: String?

  init(_ description: String) {
    errorDescription = description
  }
}

struct Settings {
  var title: String
  var text: String
  var okButton: String
  var cancelButton: String

  var width: Int?
  var height: Int?
}

struct ListOptions {
  var options = [String]()
  var defaultOption = 0
  var multiple = false
}

class TableViewController: NSViewController, NSTableViewDelegate, NSTableViewDataSource {

  var initialized = false
  var frame = NSRect()
  var opts = ListOptions()
  var defaultOption = 0
  var tableView = NSTableView()

  convenience init(frame: NSRect, options: ListOptions) {
    self.init()
    self.frame = frame
    self.opts = options
    self.tableView = NSTableView(frame: frame)
    self.view = self.tableView
  }

  override func viewDidLayout() {
    if !initialized {
      initialized = true
      setupTableView()
    }
  }

  func setupTableView() {
    self.tableView.delegate = self
    self.tableView.dataSource = self
    self.tableView.headerView = nil
    self.tableView.allowsMultipleSelection = self.opts.multiple
    self.tableView.usesAutomaticRowHeights = true
    self.tableView.backgroundColor = NSColor.clear
    self.tableView.selectionHighlightStyle = NSTableView.SelectionHighlightStyle.regular
    self.tableView.selectRowIndexes(
      IndexSet(integer: self.opts.defaultOption), byExtendingSelection: false)
    //tableView!.appearance = NSAppearance(named: NSAppearance.Name.vibrantDark)

    let col = NSTableColumn(identifier: NSUserInterfaceItemIdentifier(rawValue: "col"))
    col.minWidth = 200
    tableView.addTableColumn(col)
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

func makeDefaultAlert(_ settings: Settings, listOpts: ListOptions) -> (NSAlert, TableViewController)  {
  let a = NSAlert()
  a.messageText = settings.title
  a.alertStyle = NSAlert.Style.informational
  a.addButton(withTitle: settings.okButton)
  a.addButton(withTitle: settings.cancelButton)

  a.buttons[0].keyEquivalent = "\r"
  a.buttons[1].keyEquivalent = "\u{1b}"

  // Text as accessory view since
  // Height can not be changed
  let v = NSStackView(frame: NSRect(x: 0, y: 0, width: settings.width ?? 250, height: 350))
  v.orientation = NSUserInterfaceLayoutOrientation.vertical
  v.distribution = NSStackView.Distribution.equalSpacing
  v.spacing = 10

  // Text 1
  let text = NSTextField(wrappingLabelWithString: settings.text)
  text.isEditable = false
  text.drawsBackground = true
  v.addView(text, in: NSStackView.Gravity.center)

  // Text 2
  let text2 = NSTextField(wrappingLabelWithString: settings.text)
  text2.isEditable = false
  text2.drawsBackground = true
  v.addView(text2, in: NSStackView.Gravity.center)

  //Table
  let scrollView = NSScrollView(
    frame: NSRect(x: 0, y: 0, width: settings.width ?? 250, height: 100))
  scrollView.hasVerticalScroller = true
  scrollView.hasHorizontalScroller = true
  let clipView = NSClipView(frame: scrollView.bounds)
  clipView.autoresizingMask = [.width, .height]
  let con = TableViewController(frame: clipView.bounds, options: listOpts)
  con.tableView.autoresizingMask = [.width, .height]
  clipView.documentView = con.tableView
  scrollView.contentView = clipView
  v.addView(scrollView, in: NSStackView.Gravity.center)

  a.accessoryView = v

  return (a, con)
}

func runOptionsDialog() throws {

  let text = """
    Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore
    et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum.
    Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
    """

  let listOpts = ListOptions(options: ["option1", "option2"], multiple: true)
  let settings = Settings(title: "Title", text: text, okButton: "Ok", cancelButton: "Cancel")

  let a = makeDefaultAlert(settings, listOpts: listOpts)
  let alert = a.0
  app.activate(ignoringOtherApps: true)
  alert.runModal()
}

try runOptionsDialog()
