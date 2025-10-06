@echo off
setlocal enabledelayedexpansion

REM Установи пути
set MOCKGEN=go run github.com/golang/mock/mockgen
set DELIEVERY_PATH=gateway/internal/delivery

REM Delivery mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/http/auth-handler.go -destination=%DELIEVERY_PATH%/http/mocks/auth-mock.go -package=mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/http/feed-handler.go -destination=%DELIEVERY_PATH%/http/mocks/feed-mock.go -package=mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/http/chat-handler.go -destination=%DELIEVERY_PATH%/http/mocks/chat-mock.go -package=mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/http/csrf.go -destination=%DELIEVERY_PATH%/http/mocks/csrf-mock.go -package=mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/http/friends-handler.go -destination=%DELIEVERY_PATH%/http/mocks/friends-mock.go -package=mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/http/message-handler.go -destination=%DELIEVERY_PATH%/http/mocks/message-mock.go -package=mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/http/message-handlerWS.go -destination=%DELIEVERY_PATH%/http/mocks/messageWS-mock.go -package=mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/http/profile-handler.go -destination=%DELIEVERY_PATH%/http/mocks/profile-mock.go -package=mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/http/search-handler.go -destination=%DELIEVERY_PATH%/http/mocks/search-mock.go -package=mocks
%MOCKGEN% -source=%DELIEVERY_PATH%/ws/ws-manager.go -destination=%DELIEVERY_PATH%/ws/mocks/manager-mock.go -package=mocks

echo Mock generation completed.
