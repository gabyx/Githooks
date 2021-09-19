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

  var width = 300
  var height = 400
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
  var texts = [NSTextField]()
  var rowHeights = [CGFloat]()
  var defaultOption = 0
  var tableView = NSTableView()

  convenience init(frame: NSRect, options: ListOptions) {
    self.init()
    self.frame = frame
    self.opts = options
    self.rowHeights = Array(repeating: CGFloat(0.0), count: self.opts.options.count)
    self.tableView = NSTableView()
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

    self.tableView.style = .sourceList
    self.tableView.headerView = nil
    self.tableView.gridStyleMask = .solidHorizontalGridLineMask
    self.tableView.allowsMultipleSelection = self.opts.multiple
    self.tableView.usesAutomaticRowHeights = true
    //self.tableView.backgroundColor = NSColor.systemGray
    self.tableView.selectionHighlightStyle = NSTableView.SelectionHighlightStyle.regular
    self.tableView.selectRowIndexes(
      IndexSet(integer: self.opts.defaultOption), byExtendingSelection: false)
    //tableView!.appearance = NSAppearance(named: NSAppearance.Name.vibrantDark)

    let col = NSTableColumn(identifier: NSUserInterfaceItemIdentifier(rawValue: "col"))
    col.minWidth = self.frame.width - 30
    self.tableView.addTableColumn(col)

    for i in 0..<self.opts.options.count {
      let text = NSTextField(wrappingLabelWithString: self.opts.options[i])
      text.preferredMaxLayoutWidth = self.frame.width
      text.isEditable = false
      text.drawsBackground = false
      text.isBordered = false

      self.rowHeights[i] = max(self.rowHeights[i], text.fittingSize.height)
      print(self.rowHeights[i])
      texts.append(text)
    }
  }

  func tableView(_ tableView: NSTableView, heightOfRow: Int) -> CGFloat {
    return max(self.rowHeights[heightOfRow] + 10, 16.0)
  }

  func numberOfRows(in tableView: NSTableView) -> Int {
    return self.opts.options.count
  }

  func tableView(_ tableView: NSTableView, viewFor tableColumn: NSTableColumn?, row: Int)
    -> NSView?
  {
    let text = self.texts[row]
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

func makeDefaultAlert(_ settings: Settings, listOpts: ListOptions) -> (NSAlert, TableViewController)
{
  let a = NSAlert()
  a.messageText = settings.title
  a.alertStyle = NSAlert.Style.informational
  a.addButton(withTitle: settings.okButton)
  a.addButton(withTitle: settings.cancelButton)

  a.buttons[0].keyEquivalent = "\r"
  a.buttons[1].keyEquivalent = "\u{1b}"

  // Text as accessory view since
  // Height can not be changed
  let v = NSStackView()
  v.edgeInsets = NSEdgeInsets(top: 0, left: 0, bottom: 0,  right: 0)
  v.orientation = NSUserInterfaceLayoutOrientation.vertical
  v.distribution = NSStackView.Distribution.fill
  v.spacing = 15

  // Text 1
  let text = NSTextField(wrappingLabelWithString: settings.text)
  text.preferredMaxLayoutWidth = CGFloat(settings.width - 20)
  // text.setContentCompressionResistancePriority(.defaultLow, for: .vertical)
  text.isEditable = false
  text.isBordered = false
  text.drawsBackground = false
  v.addView(text, in: NSStackView.Gravity.center)
  print(text.fittingSize)

  // Text 2
  let text2 = NSTextField(wrappingLabelWithString: settings.text)
  text2.preferredMaxLayoutWidth = CGFloat(settings.width - 20)
  // text2.setContentCompressionResistancePriority(.defaultHigh, for: .vertical)
  text2.isBordered = false
  text2.isEditable = false
  text2.drawsBackground = false
  print(text2.fittingSize)
  v.addView(text2, in: NSStackView.Gravity.center)

  // //Table
  let sFr = NSRect(x: 0, y: 0, width: settings.width, height: min(200, 30 * listOpts.options.count))
  let scrollView = NSScrollView(
    frame: sFr)
  scrollView.hasVerticalScroller = true
  scrollView.hasHorizontalScroller = true
  let clipView = NSClipView(frame: scrollView.bounds)
  let con = TableViewController(frame: clipView.bounds, options: listOpts)
  con.tableView.sizeToFit()
  clipView.documentView = con.tableView
  scrollView.contentView = clipView
  // scrollView.setContentCompressionResistancePriority(.defaultHigh, for: .vertical)
  // scrollView.setContentHuggingPriority(.defaultLow, for: .vertical)
  v.addView(scrollView, in: NSStackView.Gravity.center)

  // v.setVisibilityPriority(.mustHold, for: text)
  // v.setVisibilityPriority(.mustHold, for: text2)
  // v.setVisibilityPriority(.mustHold, for: scrollView)
  // v.setClippingResistancePriority(.defaultHigh, for: .vertical)
  // v.setContentHuggingPriority(.defaultLow, for: .vertical)
  a.accessoryView = v

  let b = text.fittingSize.height + text2.fittingSize.height + sFr.height
  print(text.fittingSize, text2.fittingSize, sFr, b)
  v.frame = NSRect(x: 0, y: 0, width: settings.width, height: Int(b))
  a.layout()

  return (a, con)
}

func runOptionsDialog() throws {

  let text = """
    Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore
    et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum.
    Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
    """

  let listOpts = ListOptions(
    options: Array(repeating: "as as as as as ass as as as as as as as as as as as as as ", count: 30), multiple: true
  )
  let settings = Settings(title: "Title", text: text, okButton: "Ok", cancelButton: "Cancel")

  let a = makeDefaultAlert(settings, listOpts: listOpts)
  let alert = a.0
  app.activate(ignoringOtherApps: true)
  alert.runModal()
}

try runOptionsDialog()
