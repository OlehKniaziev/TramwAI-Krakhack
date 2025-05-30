import requests
from bs4 import BeautifulSoup

# Function to fetch and parse the webpage
def fetch_and_parse(url):
    # Send HTTP request to the URL
    response = requests.get(url)

    # Check if request was successful (status code 200)
    if response.status_code == 200:
        # Parse the content of the page with BeautifulSoup
        soup = BeautifulSoup(response.content, 'html.parser')
        return soup
    else:
        print(f"Failed to retrieve the page. Status code: {response.status_code}")
        return None

# Example function to scrape specific elements (change based on your needs)
def scrape_data(soup):
    # Find the title of the page
    title = soup.find('title').get_text()
    print(f"Title of the page: {title}")

    # Example: Scraping all links (anchor tags)
    links = soup.find_all('a')
    for link in links:
        href = link.get('href')
        if href:
            print(f"Link: {href}")

    # Example: Scraping all paragraphs
    paragraphs = soup.find_all('p')
    for para in paragraphs:
        print(f"Paragraph: {para.get_text()}")

# Main function to drive the script
if __name__ == '__main__':
    url = 'https://www.krakow.pl/kalendarium/1919,artykul,kalendarium.html'
    soup = fetch_and_parse(url)

    if soup:
        scrape_data(soup)
