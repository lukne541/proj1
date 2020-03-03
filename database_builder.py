import sys
import sqlite3
import os
import re
import copy


class Episode:
    def __init__(self, serie_name, season, episode, media_dir, sub_dir):
        self.serie_name = serie_name
        self.season = season
        self.episode = episode
        self.media_dir = media_dir
        self.sub_dir = sub_dir


def main(dir, db_name):
    """
    dir: directory path to the Series, ex: /home/lukas/Videos/Series/
    db_name: name of the used database, last used db was: media.db 
    """
    episodes = []
    root_dir = dir
    find_all(dir, episodes, root_dir)
    #print(episodes)

    clean_database(db_name)
    add_to_databse(db_name, episodes)

def clean_database(db_name):
    """

    :param db_name:
    :return:
    """
    conn = sqlite3.connect(db_name)
    c = conn.cursor()
    tables = []
    for table in c.execute("SELECT name FROM sqlite_master WHERE type='table'"):
        tables.append(table[0])
    for table in tables:
        c.execute("DROP TABLE IF EXISTS " + table)

def add_to_databse(db_name, episodes):
    conn = sqlite3.connect(db_name)
    c = conn.cursor()

    series = set()
    for ep in episodes:
        series.add(ep.serie_name)

    #print(series)

    for serie in series:
       # print("SERIE: " + serie)
        c.execute(r"CREATE TABLE " + serie + r" (season INTEGER, episode INTEGER, media_dir TEXT, sub_dir TEXT);")

    for ep in episodes:
        if ep.sub_dir == "": ep.sub_dir = " "
        t = (ep.season, ep.episode, ep.media_dir, ep.sub_dir)
        if ep.serie_name == "Generation_war": print(t)
        c.execute("INSERT INTO " + ep.serie_name + " VALUES (?, ?, ?, ?)", t)
    conn.commit()
    conn.close()

def find_all(dir, dict, root_dir):
    entries = os.listdir(dir)

    r = re.compile(".*\.(mp4|mkv|m4v|vtt|flac|wav)")
    filtered_entries = list(filter(r.match, entries))
    if len(filtered_entries) == 0:
        for entry in entries:
            if os.path.isdir(os.path.join(dir, entry)):
                find_all(dir+"/"+entry, dict, root_dir)
    else:
        
        for entry in filtered_entries:
            fn = entry.split("/")[-1]

            if "BBC" in entry:
                print(end="")

            if entry.endswith(".wav") or entry.endswith(".flac"):
                try:
                    season = 1
                    episode = re.search(r"[0-9]*", fn)
                except TypeError:
                    pass
            else:
                try:
                    season = re.search(r"(S|s)[0-9]{2}", fn)
                    episode = re.search(r"(E|e)[0-9]{2}", fn)
                except TypeError:
                    pass

            if entry.endswith(".wav") or entry.endswith(".flac"):
                print("wav flac", episode, season, fn)


            if episode and season and not isinstance(season, int):
                episode = episode.group(0)
                season = season.group(0)
            elif episode:
                episode = episode.group(0)

            

            try:
                if re.search(r"[0-9]", episode) is not None and re.search(r"[0-9]", season) is not None:
                    season = re.sub("[^0-9]", "", season)
                    episode = re.sub("[^0-9]", "", episode)
                else:
                    print("epi, seas", episode, season)

            except TypeError:
                print("EXCEPT epi, seas", episode, season)
                pass
                
            d = dir[len(root_dir):]
            args = d.split("/")[1:]
            sub_dir = ""
            media_dir = ""
            if entry.endswith(".m4v") or entry.endswith(".mp4") or entry.endswith(".mkv"):
                media_dir = dir + "/" + entry
                
                reg = r'.*(S|s)' + str(season) + r'.*(E|e)' + str(episode) + r'.*(.vtt)$'
                regex = re.compile(reg)
                selected = list(filter(regex.match, filtered_entries))
                if selected:
                    sub_dir = dir + "/" + selected[0]
                
                
                args[0] = args[0].replace("-", "_")
                args[0] = args[0].replace(" ", "_")
                args[0] = args[0].replace("__", "_")
                #print(args[0], " <--")

                dict.append(Episode(args[0], season, episode, media_dir, sub_dir))
            elif entry.endswith(".wav") or entry.endswith(".flac"):
                media_dir = dir + "/" + entry
                args[0] = args[0].replace("-", "_")
                args[0] = args[0].replace(" ", "_")
                args[0] = args[0].replace("__", "_")
                print(args[0])
                dict.append(Episode(args[0], str(season), str(episode), media_dir, ""))

                
if __name__=='__main__':
    try:
        main(sys.argv[1], sys.argv[2])
    except IndexError:
        main(r"/home/lukas/Videos/Series", "media.db")
