
# Welcome to Plex Collection Tool!

Plex Collection Tool enables you to purge, create and update your Plex collections based on RegEx searches or IMDB lists. Let's face it who actually goes into the Plex collections to find the movies Airplane and Airplane II, so why not put your collections to better use by having Best Movies of the 80's, 90's, 00's, Best Action Movies, Best Rom Com Movies. There are many public lists on IMDB where people have already put great effort into collating movies, so if you have these in your system already why not create a collection to make them easier to find. This is where Plex Collection Tool comes in.


# Prerequisites

You will require a Plex installation, knowledge of your servers Plex Url (not plex.tv) and Api Token.

# Usage

### Environment Variables
You can store your token and url in an environment variable if you don't want to keep it in plain text in any scripts or if you just don't want to have to type it out every time.

`PLEX_TOKEN` stores your Api token

`PLEX_URL` stores your base url

### Command line options
```
  -c string    Location of the config file, defaults to config.yml in the same folder as the application
```

### Config file

You now have the option to put everything in a configuration file so you can just run the command once without any parameters. Place a config.yml file in the same path as your executable.

You can combine multiple lists and search terms to add to one collection see the example config.yml.sample

#### Basic Example
```
config:
  plex:
    apiKey: YOUR_PLEX_API_KEY                     # Your plex token https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/
    baseURL: http://127.0.0.1:32400               # Base URL of your Plex server, not app.plex.tv
  updateDb: true                                  # Update the localdb when running, needed 
  purge: 1                                        # Remove collections with this number or fewer items in
  logging:                                        # Switch on console logging for individual types of data
    added: true                                   #   Display if Movie was added to collection
  lists:
    #Simple example of a public IMDb user list building a collection
    - name: Marvel Movies
      imdb-ids:
        - id: ls041413544
```

### Suggestions and other info
This tool works on all your Movie libraries and you can't currently specify which ones to run it on if you have more than one, this is something I will look to add in the future.

This tool uses a local database, this stores all the data of your movies in an easily queryable manner, this means it's then possible to search on the IMDb tt number to check if you have a movie and what it's unique id is in Plex.

