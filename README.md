
# Welcome to Plex Collection Tool!

Plex Collection Tool enables you to purge, create and update your Plex collections based on RegEx searches or IMDB lists. Let's face it who actually goes into the Plex collections to find the movies Airplane and Airplane II, so why not put your collections to better use by having Best Movies of the 80's, 90's, 00's, Best Action Movies, Best Rom Com Movies. There are many public lists on IMDB where people have already put great effort into collating movies, so if you have these in your system already why not create a collection to make them easier to find. This is where Plex Collection Tool comes in.


# Prerequisites

You will require a Plex installation, knowledge of your servers Plex Url (not plex.tv) and Api Token.
You will also need a MongoDb server installed, this can be done quite easily with docker.

### Install MongoDb in docker
`docker run --name PCTMongo -d --restart=always -p 27017:27017 mongo`
### Run MongoDb in the cloud
Go to https://cloud.mongodb.com/ and create a free tier which allows databases up to 512MB, this only needs about 100MB per 10,000 movies.

After creating a cluster, click connect, setup the IP whitelisting as you need, either whitelist 1 address or all addresses, then create a user, remember the password you will need it, then click Connect your application and select Node.js 2.2.12 or later this will give you a mongodb:// and not a mongodb+srv:// I will investigate why the latter is not working.

Copy the connection string and fill out the password and dbname (can be anything you like e.g. plex)
# Usage

### Environment Variables
You can store your token and url in an environment variable if you don't want to keep it in plain text in any scripts or if you just don't want to have to type it out every time.

`PLEX_TOKEN` stores your Api token

`PLEX_URL` stores your base url

`MONGO_URI` stores your MongoDb connection string Uri
### Command line options
```
  -a string    Your plex Access token
  -b string    The Base url of your plex install without trailing slash e.g. http://127.0.0.1:32400
  -c string    Name of the Collection to add titles to
  -cache       Cache http get requests, this helps when testing
  -i []string  Lists to add to collection
  -m string    MongoDb Connection String URI (default "mongodb://127.0.0.1:27017")
  -p int       Purge movie collections with less than x movies in them
  -s []string  Search term to search for
  -u           Update the local database from plex
```

### Suggestions and other info
This tool works on all your Movie libraries and you can't currently specify which ones to run it on if you have more than one, this is something I will look to add in the future.

It's probably worth updating your Movie library to disable Automatic collections, or maybe set it to 4 items or more, this way your curated lists won't disappear in a deluge of 1/2 movie collections.

This tool uses MongoDb, this stores all the data of your movies in an easily queryable database, this means it's then possible to search on the IMDb tt number to check if you have a movie and what it's unique id is in Plex.

# Examples
## Update MongoDb
Update your mongo database with information from your Plex install, this will only add new or updated information and is necessary when there is new content in your library, it can be combined with the following examples.
`pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL" -u`
## Create/Update Collection by Regex
Uses Mongo's Regex implimentation, you however can't set options, the case insensitve option is always set.

This searches for the words elf, christmas, santa and holiday and adds them to a Christmas Movies collection.

`pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL" -c "Christmas Movies" -s "\belf\b" -s "\bchristmas\b" -s "\bsanta\b" -s "\bholiday\b"`
## Create/Update Collection by IMDb lists
Create or update a collection called Christmas Movies with the contents of the IMDb list https://www.imdb.com/list/ls000096828/
`pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL" -c "Christmas Movies" -i ls000096828`

Add -u to update your MongoDb first before checking the IMDb list
`pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL" -c "Christmas Movies" -i ls000096828 -u`

Create or update a collection called Christmas Movies with the contents of two IMDb lists
`pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL" -c "Christmas Movies" -i ls000096828 -i ls006571770`
## Purge your Plex Collections
Purging collections can be slow, this is because it is waiting for the Plex server to finish the request, I will look in the future to see if it's possible to speed this up.

Probably remove all your Collections, unless you have ten thousand movies in a single collection, in which case add another 9 on the end!
`pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL" -p 99999`

Remove all your Collections that only have a single movie in them
`pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL" -p 1`
## Real world example script

In the following example the first call updates the MongoDb so it is fully up to date to search for all the movies, however it is not on the other lines to speed up their calls. It only needs to be run when new content has been added between runs of PCT.

You will notice the Comic Book Movies collection is made up of 3 IMDb lists, this is a general comic book movie list, a Marvel specific one and a DC list too.

```
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "Comic Book Movies" -i ls004135985 -i ls041413544 -i ls041927031 -u
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "Marvel Movies" -i ls041413544
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "DC Movies" -i ls041927031
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "2020 Best Movies" -i ls093785287
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "2019 Best Movies" -i ls043474895
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "2010s Best Movies" -i ls021078225
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "2000s Best Movies" -i ls000718410
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "1990s Best Movies" -i ls006658449
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "1980s Best Movies" -i ls006692819
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "Vaguely Christmas" -i ls054635542
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "Top Rom Coms" -i ls059288416
pct.exe -a "YOUR_PLEX_API_TOKEN" -b "YOUR_PLEX_URL"YOUR_PLEX_URL -c "Top Action Movies" -i ls063897780 -i ls058416162
```
