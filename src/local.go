package main

import "os"

var season = ""
var course = ""
var folder = ""

var adminF = folder + ""
var entryF = folder + ""

var message = ""
var modules = ""
var replyTo = ""

// --- x ---

var askDir = folder + ""
var logDir = folder + ""
var errDir = logDir + ""

var shDone = folder + "/shDone "
var shInit = folder + "/shInit " + adminF
var shLkup = folder + "/shLkup "
var shMail = folder + "/shMail " + adminF
var shWins = folder + "/shWins "

var logRWrite os.FileMode = 0244
var logUpdate os.FileMode = 0644
