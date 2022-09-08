
#include "LsaLogonUser.h"

void DumpToken(HANDLE hToken)
{
    WORD i; // how could we do without an i?
    BOOL           bRes;
 
    // various info buffers
 
    TOKEN_GROUPS *pGroupInfo;
    TOKEN_USER *pUserInfo;
    TOKEN_OWNER *pOwnerInfo;
    TOKEN_STATISTICS *pStatistics;
    TOKEN_PRIMARY_GROUP *pPrimaryInfo;
    TOKEN_PRIVILEGES *pPrivileges;
    TOKEN_SOURCE *pSource;
 
    // privilege variables
    
    TCHAR          szPrivilegeName[128];
    DWORD          dwPrivilegeNameLength;
    DWORD dwSize = 0;
    DWORD dwResult;
 
    // SID variables
 
    SID_NAME_USE SidType;
    LPWSTR lpName = nullptr;
    LPWSTR lpDomain = nullptr;
 
 
    pUserInfo = (TOKEN_USER *)AllocateTokenInfoBuffer(hToken,TokenUser,&dwSize);
 
    bRes = GetTokenInformation(hToken,TokenUser,pUserInfo,dwSize, &dwSize);
 
    if (FALSE == bRes)
    {
        wprintf(L"GetTokenInformation failed\n");
        ExitProcess(EXIT_FAILURE);
    }
 
    if( !LookupAccountSid(
        NULL,                      // lookup on local system
        pUserInfo->User.Sid,
        lpName,                    // buffer to recieve name
        &dwSize,
        lpDomain,
        &dwSize,
        &SidType) )
    {
         dwResult = GetLastError();
         if( dwResult == ERROR_NONE_MAPPED )
         {
             wcscpy(lpName, L"NONE_MAPPED");
         }
         else
         {
            printf("LookupAccountSid Error %u\n", GetLastError());
            exit( 0 );
         }
    }
 
    printf("User : %s\\%s\n", lpDomain, lpName);
 
    LocalFree(pUserInfo);
 
    //
    // Get Owner information
    //
 
    pOwnerInfo = (PTOKEN_OWNER) AllocateTokenInfoBuffer(hToken,TokenOwner,&dwSize);
 
    // Make the "real" call
    //
    bRes = GetTokenInformation(hToken, TokenOwner, pOwnerInfo, dwSize, &dwSize);
    if (FALSE == bRes)
    {
        printf("GetTokenInformation failed (%lui)\n",GetLastError());
        ExitProcess(EXIT_FAILURE);
    }
 
    if( !LookupAccountSid( 
        NULL,                      // lookup on local system
        pOwnerInfo->Owner,
        lpName,                    // buffer to recieve name
        &dwSize,
        lpDomain,
        &dwSize,
        &SidType ) )
    {
        dwResult = GetLastError();
        if( dwResult == ERROR_NONE_MAPPED )
            wcscpy( lpName, L"NONE_MAPPED" );
        else
        {
            printf( "LookupAccountSid Error %u\n", GetLastError() );
            exit( 0 );
        }
    }
 
    printf( "Owner : %s\\%s\n", lpDomain, lpName );
 
    LocalFree(pOwnerInfo);
 
    //
    // Now get Primary Group
    //
 
    pPrimaryInfo = (PTOKEN_PRIMARY_GROUP) AllocateTokenInfoBuffer(
        hToken,
        TokenPrimaryGroup,
        &dwSize);
 
    // Make the "real" call
    //
    bRes = GetTokenInformation(hToken,
                        TokenPrimaryGroup,
                        pPrimaryInfo,
                        dwSize, &dwSize);
    if (FALSE == bRes) {
                fprintf(stderr,"GetTokenInformation failed\n");
                ExitProcess(EXIT_FAILURE);
    }
 
    if( !LookupAccountSid( 
        NULL,                      // lookup on local system
        pPrimaryInfo->PrimaryGroup,
        lpName,                    // buffer to recieve name
        &dwSize,
        lpDomain,
        &dwSize,
        &SidType ) )
    {
        dwResult = GetLastError();
        if( dwResult == ERROR_NONE_MAPPED )
            wcscpy( lpName, L"NONE_MAPPED" );
        else
        {
            printf( "LookupAccountSid Error %u\n", GetLastError() );
            exit( 0 );
        }
    }
 
    printf( "Primary Group : %s\\%s\n", lpDomain, lpName );
 
    LocalFree(pPrimaryInfo);
 
 
    //
    // Display some of the token statistics information
    //
 
    pStatistics = (PTOKEN_STATISTICS) AllocateTokenInfoBuffer(
        hToken,
        TokenStatistics,
        &dwSize);
 
    if (FALSE == GetTokenInformation(hToken, TokenStatistics, pStatistics, dwSize, &dwSize)) 
    {
        fprintf(stderr,"GetTokenInformation failed\n");
        ExitProcess(EXIT_FAILURE);
    }
 
    printf("LUID for this instance of token [%d] [%d]\n",pStatistics->TokenId.HighPart, pStatistics->TokenId.LowPart);
    printf("LUID for this logon session      [%d] [%d]\n",pStatistics->AuthenticationId.HighPart, pStatistics->AuthenticationId.LowPart);
    
    if(pStatistics->TokenType == TokenPrimary)
        printf("Token is type PRIMARY\n");
    else
        printf("Token is type IMPERSONATION\n");
 
    //
    // Display source of token
    //
 
    pSource = (PTOKEN_SOURCE) AllocateTokenInfoBuffer(hToken,TokenSource,&dwSize);
 
 
    if (FALSE == GetTokenInformation(hToken, TokenSource, pSource, dwSize, &dwSize)) 
    {
            fprintf(stderr,"GetTokenInformation failed\n");
            ExitProcess(EXIT_FAILURE);
    }
 
    DWORD d;
    printf("Token source is ");
    WriteConsole(GetStdHandle(STD_OUTPUT_HANDLE), pSource->SourceName, 8, &d, NULL);
    printf("\n");
    //printf("Token source is <%6s>\n", pSource->SourceName );
 
    LocalFree(pSource);
 
    //
    //  List all groups
    //
 
    printf( "\nRetrieving Group information from current process token\n" );
 
    pGroupInfo = (PTOKEN_GROUPS) AllocateTokenInfoBuffer(
        hToken,
        TokenGroups,
        &dwSize);
 
    // make the real call
    if( !GetTokenInformation( hToken, TokenGroups, pGroupInfo, dwSize, &dwSize ))
    {
        dwResult = GetLastError();
        printf( "GetTokenInformation Error %u\n", dwResult );
        exit(0);
    }
 
    for(i=0;i<pGroupInfo->GroupCount;i++)
    {
       dwSize = 256;

        if(FALSE == LookupAccountSid( 
            NULL,                      // lookup on local system
            pGroupInfo->Groups[i].Sid,
            lpName,                    // buffer to recieve name
            &dwSize,
            lpDomain,
            &dwSize,
            &SidType))
        {
            dwResult = GetLastError();
            if( dwResult == ERROR_NONE_MAPPED )
                wcscpy(lpName, L"NONE_MAPPED");
            else
            {
                printf( "LookupAccountSid Error %u\n", GetLastError() );
                exit(0);
            }
        }
 
        printf( "SID %d Group: %s\\%s\n", i, lpDomain, lpName );
    }
 
    LocalFree(pGroupInfo);
 
    //
    // Display privileges associated with this token
    //
 
    pPrivileges = (PTOKEN_PRIVILEGES) AllocateTokenInfoBuffer(hToken,TokenPrivileges,&dwSize);
 
    if(GetTokenInformation( hToken, TokenPrivileges, pPrivileges, dwSize, &dwSize) == false)
    {
        fprintf(stderr,"GetTokenInformation failed\n");
        ExitProcess(EXIT_FAILURE);
    }
 
  //  Log(L"\nPrivileges associated with this token (%lu)\n",pPrivileges->PrivilegeCount );
 
    for(i=0;i<pPrivileges->PrivilegeCount;i++)
    {
        dwPrivilegeNameLength=128;
 
        if(LookupPrivilegeName(
            NULL,
            &pPrivileges->Privileges[i].Luid,
            szPrivilegeName,
            &dwPrivilegeNameLength))
 
            printf("%s - (attributes) %lu\n",szPrivilegeName, pPrivileges->Privileges[i].Attributes);
        else
            printf("LookupPrivilegeName failed - %lu\n",GetLastError());
 
    }
 
    LocalFree(pPrivileges);
 
}
 
LPVOID AllocateTokenInfoBuffer(HANDLE hToken, TOKEN_INFORMATION_CLASS InfoClass, DWORD *dwSize)
{
    BOOL bRes;
 
    *dwSize=0;
    //
    // Determine size of buffer needed
    //
 
    bRes = GetTokenInformation(
        hToken,
        InfoClass,
        NULL,
        *dwSize, dwSize);
 
    //
    // Allocate a buffer for our token data
    //
    return(LocalAlloc(LPTR,*dwSize));
}

int wmain()
{
    HANDLE hToken;
    PMSV1_0_INTERACTIVE_PROFILE Profile = NULL;
    NTSTATUS Status;
    LPWSTR wUserName = nullptr;
    LPWSTR wDomain = nullptr;
    LPWSTR wPassword = nullptr;
    BOOL KerbLogon = FALSE;

    // 
    // Prepare to use the authentication package
    // 
    Status=RegisterLogonProcess();

    if (Status != STATUS_SUCCESS)
    {
        fprintf(stderr, "RegisterLogonProcess error! (ntstatus=0x%lx)\n", Status);
        return RTN_ERROR;
    }

    // 
    // we've successfully setup a link to the authentication package
    // we can now logon via MyLogonUser until we DeRegister the
    // authentication package
    // 

    Status = MyLogonUser(&hToken, &Profile);

    if (Status == STATUS_SUCCESS)
    {

          DumpToken(hToken);

        // 
        // impersonate the user we just got a token for
        // 
        if (ImpersonateLoggedOnUser(hToken))
        {
            CHAR szName[255];
            DWORD cbName=255;

            // 
            // indicate who we logged on, to verify success
            // 
            GetUserNameA(szName, &cbName);
            fprintf(stdout,"Successfully logged on user = %s\n",
                szName);

            RevertToSelf();

            // 
            // Display the Profile
            // 
            DisplayProfile(Profile);

            // 
            // free Profile buffer allocated by Lsa
            // 
            if (Profile)
                LsaFreeReturnBuffer(Profile);
        }
        else
        {
            fprintf(stdout,"ImpersonateLoggedOnUser error (rc=%lu)\n",
            GetLastError());
        }

        // 
        // close the process token
        // 
        CloseHandle(hToken);

    }
    else
    {
        fprintf(stderr, "MyLogonUser error! (ntstatus=0x%lx)\n", Status);
    }

    // 
    // we are done with the authentication package
    // 
    DeRegisterLogonProcess();

    return RTN_OK;
}

NTSTATUS RegisterLogonProcess()
{
    CHAR szFileName[128];
    LSA_STRING LogonProcessName;
    LSA_STRING PackageName{};
    LSA_OPERATIONAL_MODE SecurityMode = 0;
    NTSTATUS Status;

    GetModuleFileNameA(NULL, szFileName, 127);

    // 
    // Prepare to use authentication services
    // 

    InitString(&LogonProcessName, szFileName);

    if ((Status=LsaConnectUntrusted(&gLsaHandle)) != STATUS_SUCCESS)
    {
        fprintf(stderr,
            "LsaConnectUntrusted failed with (%08X)\n", Status);
        return Status;
    }

    return Status;
}

void DeRegisterLogonProcess()
{
    // 
    // Close and invalidate the handle
    // 

    CloseHandle(gLsaHandle);

    gLsaHandle = INVALID_HANDLE_VALUE;
}

NTSTATUS MyLogonUser(PHANDLE pLogonToken, PMSV1_0_INTERACTIVE_PROFILE *pProfileBuffer)
{
    LSA_STRING OriginName;
    LUID LogonLuid;
    TOKEN_SOURCE SourceContext = { 0 };
    QUOTA_LIMITS Quotas;
    PMSV1_0_INTERACTIVE_LOGON MsvAuthInfo = nullptr;
    PKERB_INTERACTIVE_LOGON KerbAuthInfo = nullptr;
    ULONG ProfileBufferLength;
    PVOID AuthInfoBuf = NULL;
    ULONG AuthInfoSize = 0;
    NTSTATUS Status;
    NTSTATUS SubStatus;
    CREDUI_INFOW UiInfo = { 0 };
    DWORD dwAuthError = 0;
    ULONG ulAuthPackage = 0;
    PVOID InAuthBuf = NULL;
    ULONG InAuthBufSize = 0;
    BOOL fSave = 0;
    DWORD dwFlags = CREDUIWIN_ENUMERATE_CURRENT_USER;

    // 
    // Set logon origin
    // change to represent your product name
    // 
    char s[] = "nospam@microsoft.com";

    InitString(&OriginName, &(s[0]));

    // 
    // Initialize source context structure
    // 

   /* strncpy_s(SourceContext.SourceName, "Hello123", sizeof(SourceContext.SourceName));

    if (!AllocateLocallyUniqueId(&SourceContext.SourceIdentifier))
    {
        return STATUS_UNSUCCESSFUL;
    }*/


HDESK hDefaultDesktop = GetThreadDesktop(GetCurrentThreadId());
HDESK hWinlogonDesktop = NULL;

__try
{
    hWinlogonDesktop= OpenDesktop(L"Winlogon", 0, FALSE, MAXIMUM_ALLOWED);
    if (hWinlogonDesktop != NULL)
    {
        if (!SetThreadDesktop(hWinlogonDesktop))
        {
            fprintf(stderr, "SetThreadDesktop failed with (%08X)\n", GetLastError());
        }
        SwitchDesktop(hWinlogonDesktop);
    }

    UiInfo.cbSize = sizeof(UiInfo);
    UiInfo.pszMessageText = L"Hello";
    UiInfo.pszCaptionText = L"CredUI Prompt in Secure Desktop";
    UiInfo.hwndParent = GetDesktopWindow();

    Status = CredUIPromptForWindowsCredentialsW(&UiInfo,
                    dwAuthError,
                    &ulAuthPackage,
                    InAuthBuf,
                    InAuthBufSize,
                    &AuthInfoBuf,
                    &AuthInfoSize,
                    &fSave,
                    dwFlags);
    if (Status != STATUS_SUCCESS)
    {
        fprintf(stderr,
            "CredUIPromptForWindowsCredentials failed with (%08X)\n", Status);
        return Status;
    }

	SetThreadDesktop(hDefaultDesktop);
	SwitchDesktop(hDefaultDesktop);
	CloseDesktop(hWinlogonDesktop);
	CloseDesktop(hDefaultDesktop);

	Status = LsaLogonUser(
                    gLsaHandle,             // global LsaHandle
                    &OriginName,
                    Interactive,              // logon type
                    gAuthPackageId,         // global authentication package ID
                    AuthInfoBuf,
                    AuthInfoSize,
                    NULL,
                    &SourceContext,
                    (PVOID *)pProfileBuffer,
                    &ProfileBufferLength,
                    &LogonLuid,
                    pLogonToken,
                    &Quotas,
                    &SubStatus);

} // try
__finally
{
} // finally

return Status;
}

void DisplayProfile(PMSV1_0_INTERACTIVE_PROFILE ProfileBuffer)
{
    fprintf(stdout,   "LogonCount         = 0x%x\n",
        ProfileBuffer->LogonCount);

    fprintf(stdout,   "BadPasswordCount   = 0x%x\n",
        ProfileBuffer->BadPasswordCount);

    DisplaySystemTime("LogonTime         ", (LPFILETIME)&ProfileBuffer->LogonTime);
    DisplaySystemTime("LogoffTime        ", (LPFILETIME)&ProfileBuffer->LogoffTime);
    DisplaySystemTime("KickOffTime       ", (LPFILETIME)&ProfileBuffer->KickOffTime);
    DisplaySystemTime("PasswordLastSet   ", (LPFILETIME)&ProfileBuffer->PasswordLastSet);
    DisplaySystemTime("PasswordCanChange ", (LPFILETIME)&ProfileBuffer->PasswordCanChange);
    DisplaySystemTime("PasswordMustChange", (LPFILETIME)&ProfileBuffer->PasswordMustChange);

    fprintf(stdout,   "LogonScript        = %ls\n", ProfileBuffer->LogonScript.Buffer != NULL ? ProfileBuffer->LogonScript.Buffer : L"N/A");

    fprintf(stdout,   "HomeDirectory      = %ls\n", ProfileBuffer->HomeDirectory.Buffer != NULL ? ProfileBuffer->HomeDirectory.Buffer : L"N/A");

    fprintf(stdout,   "FullName           = %ls\n", ProfileBuffer->FullName.Buffer != NULL ? ProfileBuffer->FullName.Buffer : L"N/A");

    fprintf(stdout,   "ProfilePath        = %ls\n", ProfileBuffer->ProfilePath.Buffer != NULL ? ProfileBuffer->ProfilePath.Buffer : L"N/A");

    fprintf(stdout,   "HomeDirectoryDrive = %ls\n", ProfileBuffer->HomeDirectoryDrive.Buffer != NULL ? ProfileBuffer->HomeDirectoryDrive.Buffer : L"N/A");

    fprintf(stdout,   "LogonServer        = %ls\n", ProfileBuffer->LogonServer.Buffer != NULL ? ProfileBuffer->LogonServer.Buffer : L"N/A");

    fprintf(stdout,   "UserFlags          = 0x%lx", ProfileBuffer->UserFlags);
}

void DisplaySystemTime(LPCSTR Text,LPFILETIME ft)
{
    SYSTEMTIME utc;
    SYSTEMTIME st;

    FileTimeToSystemTime(ft, &utc);
    SystemTimeToTzSpecificLocalTime(NULL, &utc, &st);

    fprintf(stdout,"%s = %02u-%02u-%u %02u:%02u:%02u.%02u\n",
        Text, st.wMonth, st.wDay, st.wYear,
        st.wHour, st.wMinute, st.wSecond, st.wMilliseconds);
}

void InitString(PLSA_STRING LsaString, LPSTR String)
{
    if (String == NULL)
    {
        LsaString->Buffer = NULL;
        LsaString->Length = 0;
        LsaString->MaximumLength = 0;

        return;
    }

    DWORD StringLength = strlen(String);
    LsaString->Buffer = String;
    LsaString->Length = (USHORT) StringLength * sizeof(CHAR);
    LsaString->MaximumLength =
    (USHORT) (StringLength + 1) * sizeof(CHAR);
}