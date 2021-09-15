#!/usr/bin/env swift

// This script can be called by
// swift githooks/apps/dialog/gui/darwin/swift/message.swift error warning "title" "message" ok cancel 3 A B C

import AppKit
import Foundation

let app = NSApplication.shared
app.setActivationPolicy(.regular) // Magic to accept keyboard input and be docked!

let args = CommandLine.arguments

let type = args[1]

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

switch(type){
    case "message":
        runMessageDialog(args: Array<String>(args[2...]))
    case "options" :
        runOptionsDialog(args: Array<String>(args[2...]))
    default:
        die(mess:"Cannot show dialog type '\(type)'.")
}


func runMessageDialog(args: [String]) {

    let appIcon = args[1]
    let icon = args[2]
    let title = args[3]
    let text = args[4]
    let okButton = args[5]
    let cancelButton = args[6]

    let nButtons = Int(args[7]) ??  0
    var extraButtons = [String]()
    if nButtons > 0 {
        extraButtons = Array<String>(args[8..<8+nButtons])
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
    let res = a.runModal()

    if (res.rawValue >= 1000 && res.rawValue <= 1000+(nButtons+2)-1) {
        print(res.rawValue-1000)
    } else {
        exit(1)
    }

    exit(0)
}

class ViewController: NSViewController {

    @IBOutlet var tableView: NSTableView!
    var data: [[String: String]]?

    override func viewDidLoad() {
        super.viewDidLoad()

        // Do any additional setup after loading the view.
        data = [
                [
                    "firstName" : "Andrew"
                ],
                [
                    "firstName" : "Gabriel"
                ],
                [
                    "firstName" : "Olga"
            ]
        ]

        self.tableView.reloadData()
    }
}

extension ViewController: NSTableViewDataSource, NSTableViewDelegate {

    func numberOfRowsInTableView(tableView: NSTableView) -> Int {
        return (data?.count)!
    }

    func tableView(tableView: NSTableView, viewFor row: Int) -> NSView? {

        let item = (data!)[row]

        let cell = tableView.makeView(withIdentifier:NSUserInterfaceItemIdentifier("firstName"), owner: self) as? NSTableCellView
        cell?.textField?.stringValue = item["firstName"]!
        return cell
    }
}

func runOptionsDialog(args: [String]) {

    let appIcon = args[1]
    let icon = args[2]
    let title = args[3]
    let text = args[4]
    let okButton = args[5]
    let cancelButton = args[6]

    let nButtons = Int(args[7]) ??  0
    var extraButtons = [String]()
    if nButtons > 0 {
        extraButtons = Array<String>(args[8..<8+nButtons])
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

    let scrollView = NSScrollView(frame: NSRect(x: 0, y: 4, width: 200, height: 500))
    let tableView = NSTableView(frame: NSRect(x: 0, y: 2, width: 200, height: 500))
    let c = NSTableColumn(identifier: NSUserInterfaceItemIdentifier("firstName"))
    c.width = 200
    c.title = "First Name"
    scrollView.documentView = tableView
    a.accessoryView = scrollView

    let con = ViewController()
    tableView.delegate = con
    tableView.reloadData()
    // for c in ["A", "B", "C", "D"] {
    //     c.
    // }

    app.activate(ignoringOtherApps: true)
    let res = a.runModal()

    if (res.rawValue >= 1000 && res.rawValue <= 1000+(nButtons+2)-1) {
        print(res.rawValue-1000)
    } else {
        exit(1)
    }

    exit(0)
}