# bt2qbt
bt2qbt is cli tool for export from uTorrent\Bittorrent into qBittorrent (convert)
- [bt2qbt](#bt2qbt)
	- [Feature](#user-content-feature)
	- [Help](#user-content-help)
	- [Usage examples](#user-content-usage-examples)
	- [Known issuses](#user-content-known-issuses)
	
Feature:
---------
 - Processing all torrents
 - Processing torrents with subdirectories and without subdirectories
 - Processing torrents with renamed files
 - Processing torrents with non-standard encodings (for example, cp1251)
 - Processing of torrents in the not ready state *
 - Processing modified torrent names
 - Save date, metrics, status. **
 - Import of tags and labels
 - Multithreading

> [!NOTE]
> \* This torrents will not be done (0%) and will need force recheck (right click on torrent -> Force recheck)
>
> [!NOTE]
> \* If you migrate from windows to linux and use replace function attention that multiple flags -r processing one by one

> [!NOTE]
> \* If you migrate from windows to linux and yours torrent files saves to some place you must use flag --search with actual paths in yours system

> [!NOTE]
> \* If you migrate from windows to linux you may need to define path separathor with --sep flag

> [!NOTE]
>\*\* The calculation of the completed parts is based only on the priority of the files in torrent. Don't transfer global uTorrent/BitTorrent statistics.

> [!NOTE]
>\*\*\* Partially downloaded torrents will be visible as 100% completed, but in fact you will need to do a recheck (right click on torrent -> Force recheck). Without recheck torrents not will be valid. This is due to the fact that conversion of .dat files in which parts of objects are stored is not implemented.

> [!IMPORTANT]
> Don't forget before use make backup bittorrent\utorrent, qbittorrent folder. and config %APPDATA%/Roaming/qBittorrent/qBittorrent.ini. Close all this program before.
> 
> [!IMPORTANT]
> You must previously disable option "Append .!ut/.!bt to incomplete files" in preferences of uTorrent/Bittorrent, or that files will not handled

Help:
-------

Help (from cmd or powershell)

```
Usage:
  C:\Users\user\Downloads\bt2qbt.exe [OPTIONS]

Application Options:
  -s, --source=         Source directory that contains resume.dat and torrents
                        files (default:
                        C:\Users\user\AppData\Roaming\uTorrent\)
  -d, --destination=    Destination directory BT_backup (as default) (default:
                        C:\Users\user\AppData\Local\qBittorrent\BT_backup\)
  -c, --config=         qBittorrent config file (for write tags) (default:
                        C:\Users\user\AppData\Roaming\qBittorrent\qBittorren-

                        t.ini)
      --without-labels  Do not export/import labels
      --without-tags    Do not export/import tags
  -t, --search=         Additional search path for torrents files
                        Example: --search='/mnt/olddisk/savedtorrents'
                        --search='/mnt/olddisk/workstorrents'
  -r, --replace=        Replace paths.
                        Delimiter for from/to is comma - ,
                        Example: -r "D:\films,/home/user/films" -r
                        "D:\music,/home/user/music"

      --sep=            Default path separator that will use in all paths. You
                        may need use this flag if you migrating from windows to
                        linux in some cases (default: \)

Help Options:
  -h, --help            Show this help message

```

Usage examples:
----------------

- If you just run application, it will processing torrents from %APPDATA%\uTorrent\ to %LOCALAPPDATA%\qBittorrent\BT_BACKUP\
```
C:\Users\user\Downloads> .\bt2qbt.exe
It will be performed processing from directory C:\Users\user\AppData\Roaming\uTorrent\ to directory C:\Users\user\AppData\Local\qBittorrent\BT_backup\
Check that the qBittorrent is turned off and the directory C:\Users\user\AppData\Local\qBittorrent\BT_backup\ and config C:\Users\user\AppData\Roaming\qBittorrent\qBittorrent.ini is backed up.
Check that you previously disable option "Append .!ut/.!bt to incomplete files" in preferences of uTorrent/Bittorrent 


Press Enter to start

Started
1/2 Sucessfully imported 1.torrent
2/2 Sucessfully imported 2.torrent

Press Enter to exit
```

- Run application from cmd or powershell with keys, if you want change source dir or destination dir, or export/import behavior
```
C:\Users\user\Downloads> .\bt2qbt.exe -s C:\Users\user\AppData\Roaming\BitTorrent\
It will be performed processing from directory C:\Users\user\AppData\Roaming\BitTorrent\ to directory C:\Users\user\AppData\Local\qBittorrent\BT_backup\
Check that the qBittorrent is turned off and the directory C:\Users\user\AppData\Local\qBittorrent\BT_backup\ is backed up.
Check that you previously disable option "Append .!ut/.!bt to incomplete files" in preferences of uTorrent/Bittorrent 


Press Enter to start
Started
1/3233 Sucessfully imported 1.torrent
2/3233 Sucessfully imported 2.torrent
3/3233 Sucessfully imported 3.torrent
...
3231/3233 Sucessfully imported 3231.torrent
3232/3233 Sucessfully imported 3232.torrent
3233/3233 Sucessfully imported 3233.torrent

Press Enter to exit
```
Known issuses:
---------------
 - Unknown
