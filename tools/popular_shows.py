import re
from bs4 import BeautifulSoup

shows = []
# The file should contain the contents of the div with id 'main' of http://www.imdb.com/search/title?start=0&title_type=tv_series
# The entry page breaks soup.
for f in ['part1.html', 'part2.html']:
    s = BeautifulSoup(open("/tmp/" + f).read())
    for td in s.find_all("td", class_="title"):
        shows = shows + [td.find_next(href=re.compile("title")).text]

out = open("../sources/shows.go", 'w')

out.write("package sources\n\n")
out.write("var shows = [...]string {\n")
for s in shows:
    out.write("    \"" + s + "\",\n")
out.write("}")

out.close()
