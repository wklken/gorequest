GoRequest Changelog
=========

## GoRequest v1.0.1 (2022-02-19)

### BUGFIXES

- forceType not working while it's only be changed in getResponseBytes https://github.com/wklken/gorequest/pull/25

### ENHANCEMENTS

- add: UserAgent(), set user-agent by s.UserAgent("") https://github.com/wklken/gorequest/pull/20
- add: Stats, collecte statistics for SuperAgent request https://github.com/wklken/gorequest/pull/21
- add: DisableCompression() https://github.com/wklken/gorequest/pull/22
- add: Mock() support HTTP mocking https://github.com/wklken/gorequest/pull/23
- add: Timeouts() support http client timeout details https://github.com/wklken/gorequest/pull/27
- enable custom Content-Type for SendFile https://github.com/wklken/gorequest/pull/26

### OTHERS

- upgrade safeModifyTransport() copy transport to support go 1.16 https://github.com/wklken/gorequest/pull/24

## GoRequest v1.0.0 (2022-01-19)

### BUGFIXES

- send `{}`: https://github.com/wklken/gorequest/pull/6
- converting query keys to lower case: https://ggithub.com/wklken/gorequest/pull/7
- query with int64 will lose accuracy: https://github.com/wklken/gorequest/pull/8
- HTTP Get not set s.TargetType to "json" https://github.com/wklken/gorequest/pull/11
- fix retry if errors: https://github.com/wklken/gorequest/pull/13

### ENHANCEMENTS

- add: wrap the EndStruct error if json decode fail: https://github.com/wklken/gorequest/pull/9
- add docs, string bytes conv: https://github.com/wklken/gorequest/pull/10
- add context support: https://github.com/wklken/gorequest/pull/12
- add: SetHeaders: https://github.com/wklken/gorequest/pull/14

### OTHERS

- upgrade go to 1.16
- add go mod support

=====================================

previous changelog of origin repo

v0.2.15 (2016-08-30)

	Features
		* Allow float and boolean in Query()'s queryStruct @davyzhang
		* Support Map in Query() @yangmls
		* Support Map in Send() @longlongh4
		* Document RedirectPolicy @codegoalie
		* Enable Debug mode by ENV variable @heytitle
		* Add Retry() @xild
	Bug/Fixes
		* Allow bodies with all methods @pkopac
		* Change std "errors" pkg to "github.com/pkg/errors" @pkopac

v0.2.14 (2016-08-30)

	Features
		* Support multipart @fraenky8
		* Support OPTIONS request @coderhaoxin
		* Add EndStruct method @franciscocpg
		* Add AsCurlCommand @jaytaylor
		* Add CustomMethod @WaveCutz
		* Add MakeRequest @quangbuule
	Bug/Fixes
		* Disable keep alive by default


v0.2.13 (2015-11-21)

	Features
		* Add DisableTransportSwap to stop gorequest from modifying Transport settings.
			Note that this will effect many functions that modify gorequest's
			tranport. (So, use with caution.) (Bug #47, PR #59 by @piotrmiskiewicz)


v0.2.12 (2015-11-21)

	Features
		* Add SetCurlCommand for printing comparable CURL command of the request
		(PR #60 by @QuentinPerez)

v0.2.11 (2015-10-24)

	Bug/Fixes
		* Add support to Slice data structure (Bug #40, #42)
		* Refactor sendStruct to be public function SendStruct

v0.2.10 (2015-10-24)

	Bug/Fixes
		* Fix incorrect text output in tests (PR #52 by @QuentinPerez)
		* Fix Panic and runtime error properly (PR #53 by @QuentinPerez)
		* Add support for "text/plain" and "application/xml" (PR #51 by
		@smallnest)
		* Content-Type header is also equivalent with Type function to identify
		supported Gorequest's Target Type

v0.2.9 (2015-08-16)

	Bug/Fixes
		* ParseQuery accepts ; as a synonym for &. thus Gorequest Query won't
		accept ; as in a query string. We add additional Param to solve this  (PR
		#43 by @6david9)
		* AddCookies for bulk adding cookies (PR #46 by @pencil001)

v0.2.8 (2015-08-10)

  Bug/Fixes
    * Added SetDebug and SetLogger for debug mode (PR #28 by @dafang)
    * Ensure the response Body is reusable (PR #37 by alaingilbert)

v0.2.7 (2015-07-11)

	Bug/Fixes
		* Incorrectly reset "Authentication" header (Hot fix by @na-ga PR #38 & Issue #39)

v0.2.6 (2015-07-10)

  Features
    * Added EndBytes (PR #30 by @jaytaylor)

v0.2.5 (2015-07-01)

  Features
    * Added Basic Auth support (pull request #24 by @dickeyxxx)

  Bug/Fixes
    * Fix #31 incorrect number conversion (PR #34 by @killix)

v0.2.4 (2015-04-13)

	Features
		* Query() now supports Struct as same as Send() (pull request #25 by @figlief)

v0.2.3 (2015-02-08)

	Features
  	* Added Patch HTTP Method

	Bug/Fixes
		* Refactored testing code

v0.2.2 (2015-01-03)

	Features
  	* Added TLSClientConfig for better control over tls
		* Added AddCookie func to set "Cookie" field in request (pull request #17 by @austinov) - Issue #7
		* Added CookieJar (pull request #15 by @kemadz)

v0.2.1 (2014-07-06)

	Features
  	* Implemented timeout test

	Bugs/Fixes
  	* Improved timeout feature by control over both dial + read/write timeout compared to previously controlling only dial connection timeout.

v0.2.0 (2014-06-13) - Send is now supporting Struct type as a parameter

v0.1.0 (2014-04-14) - Finished release with enough rich functions to do get, post, json and redirectpolicy

