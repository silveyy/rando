#pragma once
#pragma warning(disable : 4996)
#pragma comment(lib, "secur32.lib")
#pragma comment(lib, "advapi32.lib")
#pragma comment(lib, "credui.lib")
#pragma comment(lib, "user32.lib")
#pragma comment(lib, "ole32.lib")

#include <windows.h>
#include <stdio.h>
#include <ntsecapi.h>
#include <tchar.h>
#include <wincred.h>
#include <strsafe.h>

#define RTN_OK 0
#define RTN_USAGE 1
#define RTN_ERROR 13

// 
// if you have the ddk, include ntstatus.h
// 
#ifndef STATUS_SUCCESS
	#define STATUS_SUCCESS                  ((NTSTATUS)0x00000000L)
	#define STATUS_UNSUCCESSFUL             ((NTSTATUS)0xC0000001L)
#endif

// 
// cache LsaHandle and the authentication package Id
// 
HANDLE gLsaHandle = INVALID_HANDLE_VALUE;
ULONG gAuthPackageId = 0;

LPVOID AllocateTokenInfoBuffer(HANDLE hToken, TOKEN_INFORMATION_CLASS InfoClass, DWORD* dwSize);
NTSTATUS RegisterLogonProcess(void);
NTSTATUS MyLogonUser(PHANDLE pLogonToken, PMSV1_0_INTERACTIVE_PROFILE* pProfileBuffer);
void DeRegisterLogonProcess(void);
void DisplayProfile(PMSV1_0_INTERACTIVE_PROFILE ProfileBuffer);
void DisplaySystemTime(LPCSTR Text, LPFILETIME ft);
void InitString(PLSA_STRING LsaString, LPSTR String);