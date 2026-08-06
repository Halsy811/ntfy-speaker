Option Explicit
On Error Resume Next

Dim WshShell, FSO, WMI, WshNetwork
Dim args, i, arg, rawArgs, topicName, HostName
Dim TempFilePath, SourceFile, DestFile, ExeName
Dim colProcesses, proc, isRunning, RunCommand
Dim searchStr

Set WshShell = CreateObject("WScript.Shell")
Set FSO = CreateObject("Scripting.FileSystemObject")
Set WMI = GetObject("winmgmts:\\.\root\cimv2")
Set WshNetwork = CreateObject("WScript.Network")

' Get PC name and convert to lowercase
HostName = WshNetwork.ComputerName
topicName = LCase(HostName)

' Read arguments passed from Task Scheduler (like -p and -s)
Set args = WScript.Arguments
rawArgs = ""

For i = 0 To args.Count - 1
    arg = args(i)
    If InStr(arg, " ") > 0 Then
        rawArgs = rawArgs & Chr(34) & arg & Chr(34) & " "
    Else
        rawArgs = rawArgs & arg & " "
    End If
Next

rawArgs = Trim(rawArgs)

' Append the dynamic topic name to the arguments
If rawArgs <> "" Then
    rawArgs = rawArgs & " -t " & topicName
Else
    rawArgs = "-t " & topicName
End If

' This is what we will search for in WMI to identify OUR specific process
searchStr = "-t " & topicName

' Setup paths
ExeName = "ntfy-speaker.exe"
TempFilePath = WshShell.ExpandEnvironmentStrings("%tmp%")
SourceFile = "\\sharedDir\" & ExeName
DestFile = FSO.BuildPath(TempFilePath, ExeName)

' Infinite watchdog loop
Do
    Set colProcesses = WMI.ExecQuery("Select * from Win32_Process Where Name = '" & ExeName & "'")
    isRunning = False
    
    For Each proc In colProcesses
        If Not IsNull(proc.CommandLine) Then
            If InStr(1, proc.CommandLine, searchStr, vbTextCompare) > 0 Then 
                isRunning = True
            End If
        End If
    Next

    ' If our specific listener is not running, copy and start it
    If isRunning = False Then
        FSO.CopyFile SourceFile, DestFile, True
        RunCommand = Chr(34) & DestFile & Chr(34) & " " & rawArgs
        
        ' 0 = hidden, False = async (do not wait for exit)
        WshShell.Run RunCommand, 0, False
        
        ' Wait 5 secs to let the process initialize before next check
        WScript.Sleep 5000
    End If

    ' Check every 10 seconds
    WScript.Sleep 10000
Loop