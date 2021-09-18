<p align="center">
  <a href="https://www.google.com/chrome/index.html">
    <img width="20%" alt="FixChromeLauncher" src="https://upload.wikimedia.org/wikipedia/commons/a/a5/Google_Chrome_icon_%28September_2014%29.svg">
  </a>
</p>

# Fix Chrome Launcher
A Windows service to keep Chrome from destroying custom launch options.

## Summary
Here's the simple truth: I only browse with Chrome's incognito mode.  Whether or not I can actually trust Google to do what it says, I don't like keeping
cookies in my browser, so I never allow them to stick.  Even after I close "secure" browser windows, I still go to the Chrome settings and "Clear browsing data."  Yeah, I'm _that_ anal about it.

However, clicking (or right-clicking) links anywhere triggers the use of a specific launcher string for the Chrome browser.  This string can be found
in the Windows registry under the path

`HKEY_CLASSES_ROOT\ChromeHTML\shell\open\command`

The default entry for this key holds the command that Windows will execute when following links you click.

Great!  I'll just insert my custom option ("--incognito") into this string, and indirect browsing will always use that mode.  Right?

Wrong.  Every time Chrome does any kind of update, it resets this string, and suddenly I'm opening browser windows in standard "tracking" mode.

## Let's fix this stupidity

"Fix Chrome Launcher" attaches itself to this registry key like stink on...excrement.  It keeps a watchful eye on the value of the string, and instantly (well, almost instantly) restores the custom options to that launcher command that you so painstakingly troubled to put there.

The custom options you want to "stick" in the launcher string will reside in the same registry key.  A new REG_SZ value called `fcl_options` will hold the options (and their attendant args) that need to be reapplied.  These values will be in JSON format, and options without args will use an empty string for that value.

As an example, for my particular needs, this registry string value would look like

`{"--incognito" : ""}`

Optionally, you can add another entry to this registry key called `fcl_interval`.  This is a REG_DWORD value that holds the specific polling interval you would like the service to use when checking the integrity of the launcher string.  By default, the service checks every 60 seconds. You can override this with a new value in this integer entry.

I did some checking, and Chrome does not delete its primary registry key when it updates (or when it installs for the first time, for that matter), so these custom key values should be safe and persistent between updates of the browser.

## Building

On Windows, compile with: `go generate & go build -ldflags "-s -w" .`

The service uses the excellent `github.com/kardianos/service` module to perform its operations.  All dependencies should be retrieved by the Go system as part of the build process.

## Installing

Since this is a Windows Service, you will need to open a command window (with elevated privileges) to install or uninstall it.

To install:

`fix_chrome_launcher.exe -service install`

To uninstall (make sure it is not running!):

`fix_chrome_launcher.exe -service uninstall`

## Running

You can start/stop the service from the command line, or you can manage it from the Windows Services panel.  Operating from the "Local System" account should be sufficient; in my testing, it has sufficient privileges to read the key's values, and modify the default key value.

The service emits messages to the system console, so you can check there (i.e., `Event Viewer` -> `Windows Logs` -> `Application`) for any runtime error messages.  Look for the SourceName "FixChromeLauncher".

I hope you find this useful.