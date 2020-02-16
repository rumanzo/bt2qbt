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
 - Save date, metrics, status. **
 - Import of tags and labels
 - Multithreading

> [!NOTE]
> \* This torrents will not be done (0%) and will need force rehash

> [!NOTE]
>\*\* The calculation of the completed parts is based only on the priority of the files in torrent

> [!NOTE]
>\*\*\* Partially downloaded torrents will be visible as 100% completed, but in fact you will need to do a rehash. Without rehash torrents not will be valid. This is due to the fact that conversion of .dat files in which parts of objects are stored is not implemented.

> [!IMPORTANT]
> Don't forget before use make backup bittorrent\utorrent, qbittorrent folder. and config %APPDATA%/Roaming/qBittorrent/qBittorrent.ini. Close all this program before.

Help:
-------

Help (from cmd or powerwhell)

```
C:\Users\user\Downloads> .\bt2qbt_v1.0_amd64.exe -h
Usage of C:\Users\user\Downloads\bt2qbt_v1.0_amd64.exe:
-c, --qconfig (= "C:\\Users\\user\\AppData\\Roaming\\qBittorrent\\qBittorrent.ini")
    qBittorrent config files (for write tags)
-d, --destination (= "C:\\Users\\user\\AppData\\Local\\qBittorrent\\BT_backup\\")
    Destination directory BT_backup (as default)
--replace (= "")
    Replace paths.
        Delimiter for replaces - ;
        Delimiter for from/to - ,
        Example: "D:\films,/home/user/films;\,/"
        If you use path separator different from you system, declare it mannually

-s, --source (= "C:\\Users\\user\\AppData\\Roaming\\uTorrent\\")
    Source directory that contains resume.dat and torrents files
--without-labels  (= false)
    Do not export/import labels
--without-tags  (= false)
    Do not export/import tags
```

Usage examples:
----------------

- If you just run application, it will processing torrents from %APPDATA%\uTorrent\ to %LOCALAPPDATA%\qBittorrent\BT_BACKUP\
```
C:\Users\user\Downloads> .\bt2qbt_v1.0_amd64.exe
It will be performed processing from directory C:\Users\user\AppData\Roaming\uTorrent\ to directory C:\Users\user\AppData\Local\qBittorrent\BT_backup\
Check that the qBittorrent is turned off and the directory C:\Users\user\AppData\Local\qBittorrent\BT_backup\ and config C:\Users\user\AppData\Roaming\qBittorrent\qBittorrent.ini is backed up.


Press Enter to start

Started
1/2 Sucessfully imported 1.torrent
2/2 Sucessfully imported 2.torrent

Press Enter to exit
```

- Run application from cmd or powershell with keys, if you want change source dir or destination dir, or export/import behavior
```
C:\Users\user\Downloads> .\bt2qbt_v1.0_amd64.exe -s C:\Users\user\AppData\Roaming\BitTorrent\
It will be performed processing from directory C:\Users\user\AppData\Roaming\BitTorrent\ to directory C:\Users\user\AppData\Local\qBittorrent\BT_backup\
Check that the qBittorrent is turned off and the directory C:\Users\user\AppData\Local\qBittorrent\BT_backup\ is backed up.

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