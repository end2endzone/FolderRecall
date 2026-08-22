@echo off

powershell -nologo -executionpolicy bypass -File "%~dpn0.ps1"
IF %ERRORLEVEL% NEQ 0 echo An error occurred during execution of script %~n0.ps1. Error level is %ERRORLEVEL%. && EXIT /B %ERRORLEVEL%
