#!/usr/bin/env swift

// This script can be called by
// swift githooks/apps/dialog/gui/darwin/swift/message.swift error warning "title" "message" ok cancel 3 A B C

import AppKit
import Foundation

let app = NSApplication.shared
app.setActivationPolicy(.regular) // Magic to accept keyboard input and be docked!

let args = CommandLine.arguments

let iconDefault = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertNoteIcon.icns"
let errorPath =  "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns"
let warningPath = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertCautionIcon.icns"
let questionPath = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/GenericQuestionMarkIcon.icns"

final class StandardErrorOutputStream: TextOutputStream {
    func write(_ string: String) {
        FileHandle.standardError.write(Data(string.utf8))
    }
}

var outputStream = StandardErrorOutputStream()

func die(mess: String) {
    print(mess, to: &outputStream)
    exit(1)
}

func getIcon(what: String) -> NSImage? {
    switch(what) {
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

func runMessageDialog(args: [String]) {

    let appIcon = args[0]
    let icon = args[1]
    let title = args[2]
    let text = args[3]
    let okButton = args[4]
    let cancelButton = args[5]

    var i = 6
    let nButtons = Int(args[i]) ?? 0
    i+=1
    var extraButtons = [String]()
    if nButtons > 0 {
        extraButtons = Array<String>(args[i..<i+nButtons])
        i+=nButtons
    }

    // Set dock icon
    app.dockTile.contentView = NSImageView(image: getIcon(what: appIcon)!)
    app.dockTile.display()

    let a = NSAlert()
    a.messageText = title
    a.icon = getIcon(what: icon)!
    a.alertStyle = NSAlert.Style.warning
    a.informativeText = text
    a.addButton(withTitle: okButton)
    a.addButton(withTitle: cancelButton)

    a.buttons[0].keyEquivalent = "\r"
    a.buttons[1].keyEquivalent = "\u{1b}"

    for extra in extraButtons {
        a.addButton(withTitle: extra)
    }

    app.activate(ignoringOtherApps: true)
    var r = a.runModal().rawValue

    r -= 1000
    if ( r >= 0 && r <= nButtons+2-1) {
        print(r)
        exit(0)
    }

    exit(1)
}


class TableViewController: NSViewController, NSTableViewDelegate, NSTableViewDataSource {

    var initialized = false
    var frame = NSRect()
    var options: [String] = []
    var defaultOption = 0
    var tableView: NSTableView? = nil

    convenience init(frame: NSRect, options: [String], defaultOption: Int) {
        self.init()
        self.frame = frame
        self.options = options
        self.defaultOption = defaultOption
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
        self.tableView!.allowsMultipleSelection = true
        self.tableView!.usesAutomaticRowHeights = true
        self.tableView!.backgroundColor = NSColor.clear
        self.tableView!.selectionHighlightStyle =  NSTableView.SelectionHighlightStyle.regular
        self.tableView!.selectRowIndexes(IndexSet(integer: self.defaultOption), byExtendingSelection: false)
        //tableView!.appearance = NSAppearance(named: NSAppearance.Name.vibrantDark)

        let col = NSTableColumn(identifier: NSUserInterfaceItemIdentifier(rawValue: "col"))
        col.minWidth = 200
        tableView!.addTableColumn(col)
    }

    func numberOfRows(in tableView: NSTableView) -> Int {
        return self.options.count
    }

    func tableView(_ tableView: NSTableView, viewFor tableColumn: NSTableColumn?, row: Int) -> NSView? {
        let text = NSTextField()
        text.stringValue = self.options[row]
        text.isEditable = false
        text.drawsBackground = false
        text.isBordered = false
        let cell = NSTableCellView()
        cell.addSubview(text)

        text.translatesAutoresizingMaskIntoConstraints = false
        cell.addConstraint(NSLayoutConstraint(item: text, attribute: .centerY, relatedBy: .equal, toItem: cell, attribute: .centerY, multiplier: 1, constant: 0))
        cell.addConstraint(NSLayoutConstraint(item: text, attribute: .left, relatedBy: .equal, toItem: cell, attribute: .left, multiplier: 1, constant: 0))
        cell.addConstraint(NSLayoutConstraint(item: text, attribute: .right, relatedBy: .equal, toItem: cell, attribute: .right, multiplier: 1, constant: 0))

        return cell
    }

    func tableView(_ tableView: NSTableView, rowViewForRow row: Int) -> NSTableRowView? {
        let rowView = NSTableRowView()
        rowView.isEmphasized = true
        return rowView
    }
}

func runOptionsDialog(args: [String]) {

    let appIcon = args[0]
    let icon = args[1]
    let title = args[2]
    let text = args[3]
    let okButton = args[4]
    let cancelButton = args[5]

    var i = 6

    let nOptions = Int(args[i]) ??  0
    i+=1
    var options = [String]()
    if nOptions > 0 {
        options = Array<String>(args[i..<i+nOptions])
        i+=nOptions
    }

    let defaultOption = Int(args[i]) ?? 0

    // Set dock icon
    app.dockTile.contentView = NSImageView(image: getIcon(what: appIcon)!)
    app.dockTile.display()

    let a = NSAlert()
    a.messageText = title
    a.icon = getIcon(what: icon)!
    a.alertStyle = NSAlert.Style.warning
    a.informativeText = text
    a.addButton(withTitle: okButton)
    a.addButton(withTitle: cancelButton)

    a.buttons[0].keyEquivalent = "\r"
    a.buttons[1].keyEquivalent = "\u{1b}"

    let scrollView = NSScrollView(frame: NSRect(x:0, y:0, width: 400, height:100))
    scrollView.hasVerticalScroller = true
    scrollView.hasHorizontalScroller = true
    let clipView = NSClipView(frame: scrollView.bounds)
    clipView.autoresizingMask = [.width, .height]

    let con = TableViewController(frame: clipView.bounds, options: options, defaultOption: defaultOption)
    let tableView = con.tableView
    tableView!.autoresizingMask = [.width, .height]
    clipView.documentView = tableView
    scrollView.contentView = clipView
    a.accessoryView = scrollView

    app.activate(ignoringOtherApps: true)
    let r = a.runModal().rawValue

    if (r == 1000) {
        print(con.tableView!.selectedRowIndexes.map({"\($0)"}).joined(separator: ","))
    } else {
        exit(1)
    }

    exit(0)
}

// let type = args[1]
// switch(type){
//     case "message":
//         runMessageDialog(args: Array<String>(args[2...]))
//     case "options" :
//         runOptionsDialog(args: Array<String>(args[2...]))
//     default:
//         die(mess:"Cannot show dialog type '\(type)'.")
// }


runOptionsDialog(args:["info", "warning", "title", "whast", "ok", "cancel", "3", "A", "B", "Casd lkajsdl kajsdlk jaslkdj kjashd kjahsd lkasjd lkajsdl kjasld kja lsdkj  ", "1"])