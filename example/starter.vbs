Option Explicit
On Error Resume Next

Dim WshShell, FSO, WMI
Dim args, i, arg, rawArgs, topicName
Dim TempFilePath, SourceFile, DestFile, ExeName
Dim colProcesses, proc, isRunning, RunCommand
Dim searchStr, searchStrQuoted

Set WshShell = CreateObject("WScript.Shell")
Set FSO = CreateObject("Scripting.FileSystemObject")
Set WMI = GetObject("winmgmts:\\.\root\cimv2")

Set args = WScript.Arguments
If args.Count = 0 Then WScript.Quit 1

rawArgs = ""
topicName = ""

For i = 0 To args.Count - 1
    arg = args(i)
    If InStr(arg, " ") > 0 Then
        rawArgs = rawArgs & Chr(34) & arg & Chr(34) & " "
    Else
        rawArgs = rawArgs & arg & " "
    End If
    If LCase(arg) = "-t" And i + 1 < args.Count Then
        topicName = args(i + 1)
    End If
Next

rawArgs = Trim(rawArgs)
If topicName = "" Then WScript.Quit 1

searchStr = "-t " & topicName
searchStrQuoted = "-t " & Chr(34) & topicName & Chr(34)

ExeName = "ntfy-speaker.exe"
TempFilePath = WshShell.ExpandEnvironmentStrings("%tmp%")
SourceFile = "\\sharedDir\" & ExeName
DestFile = FSO.BuildPath(TempFilePath, ExeName)

Do
    Set colProcesses = WMI.ExecQuery("Select * from Win32_Process Where Name = '" & ExeName & "'")
    isRunning = False
    For Each proc In colProcesses
        If Not IsNull(proc.CommandLine) Then
            If InStr(1, proc.CommandLine, searchStr, vbTextCompare) > 0 Then isRunning = True
            If InStr(1, proc.CommandLine, searchStrQuoted, vbTextCompare) > 0 Then isRunning = True
        End If
    Next

    If isRunning = False Then
        FSO.CopyFile SourceFile, DestFile, True
        RunCommand = Chr(34) & DestFile & Chr(34) & " " & rawArgs
        WshShell.Run RunCommand, 0, False
        WScript.Sleep 5000
    End If

    WScript.Sleep 10000
Loop