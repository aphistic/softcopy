# Thoughts

Various thoughts while working on papertrail stuff...

## Documents
All documents start out with the title as the file they're uploaded as.  When going through the webui "Add Document" it should just be a "file upload" dialog that will upload a document to the Inbox the same way a file uploaded via FTP would be.  This keeps the document management consistent no matter where the document is coming from.

## Search
To start, the search page will only be able to search on tags.  When typing in the search box the user should be presented with a list of tags matching the text, then be added to the search box as one of those "token" type pill things.  Tag searches are "AND"ed together.

## Inbox
The Inbox is where documents that have just been uploaded reside.  Any uploaded files will have the system tag "unfiled" applied to them and will continue showing up in the Inbox until that tag has been removed from the document.  This allows users to add preliminary tags to a document but have it still show up in the Inbox until they're ready for it to be removed.

## Tag Types
A tag, when added, does not have a type defined and shows up with a grey background.  Tags can be modified, though, to have a type that will dictate the color the tag will show up as within the app.  For example, there could be a "Business" tag type that is green and any tag with that type would show up with a green background to make it easier to distinguish what tags mean.

## Database and Filesystem
The database will probably be SQLite to store metadata and standard files on the filesystem to store documents.  When a document is uploaded it'll be assigned a uuid and that's what it will be stored under in the filesystem.  My goal is to have all data under the same directory so it's easy to back up and restore.  An instance of papertrail should be able to be recovered just by placing backed up data in the data directory and starting the server.