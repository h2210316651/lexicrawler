# LexiCrawler - Your LLM-Ready Web Content Harvester

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**LexiCrawler** is a powerful Go-based web crawling API meticulously designed to extract, clean, and transform web page content into a pristine Markdown format, perfectly tailored for Large Language Models (LLMs).  Stop feeding your LLMs messy HTML – start giving them the clear, structured text they crave with LexiCrawler!

---

## ✨ Key Features - Supercharge Your LLM Data Pipeline

LexiCrawler isn't just another web crawler; it's a content refinement engine built for the AI era.  Here's what makes it stand out:

*   **📝 LLM-Optimized Markdown Output:**  Delivers content in clean, well-formatted Markdown, the ideal input for optimal LLM performance. Say goodbye to HTML parsing headaches in your AI workflows.

*   **📖 Intelligent Readability Enhancement:**  Powered by `go-shiori/go-readability`, LexiCrawler expertly strips away website clutter – navigation, ads, sidebars – focusing on the core, readable article content. Maximize the signal, minimize the noise for your models.

*   **▶️ Dynamic Content Mastery with JavaScript Rendering:**  Utilizing `chromedp`, LexiCrawler conquers modern web pages.  It executes JavaScript, ensuring you capture dynamically loaded content that static scrapers miss. No page is out of reach!

*   **🕷️ Efficient & Configurable Web Crawling:** Built on `gocolly/colly`, LexiCrawler offers robust, asynchronous crawling. Define allowed domains, set crawl depths, and respect robots.txt – all with Go speed and efficiency.

*   **Essential Metadata Extraction:**  Automatically extracts crucial metadata like page titles and descriptions, providing valuable context alongside the content for richer LLM understanding.

*   **Structured Data Snippets (Example Included):**  Demonstrates the power to extract structured information. The included example extracts blog post titles and links, showcasing the potential for tailored data harvesting.

*   **📸 Optional Screenshot Capture:**  Need visual documentation?  LexiCrawler can capture screenshots of crawled pages, providing a visual record alongside the text content.

*   **📦 Smart Content Caching:**  Reduces redundant crawling and speeds up development with built-in in-memory caching. Get faster iterations and save on network resources.

*   **Basic Heuristics Filtering:**  Includes initial heuristics to filter out very short paragraphs, further refining content quality and focusing on substantial text.

*   **⚙️ Highly Configurable:** Tailor LexiCrawler to your specific needs with a comprehensive configuration:
    *   Target URL and Allowed Domains
    *   Maximum Crawl Depth
    *   JavaScript Execution Control
    *   Screenshotting Toggle
    *   Caching Enable/Disable
    *   Readability Feature Switch
    *   Heuristics On/Off

*   **🔌 Simple REST API Interface:**  Exposed as a straightforward REST API using `gofiber`, making integration into your existing applications and data pipelines effortless. Just send a URL and receive clean Markdown!

*   **🚀 Built with Go Performance:**  Leverage the speed, concurrency, and efficiency of the Go programming language for rapid and scalable web crawling.

---

## 🚀 Getting Started - Crawl in Minutes

Ready to unleash LexiCrawler? Follow these simple steps:

### Prerequisites

*   **Go Installation:** Ensure you have Go installed on your system. Download from [https://go.dev/dl/](https://go.dev/dl/).

### Installation

1.  **Get the package:**

    ```bash
    go get -u -v github.com/h2210316651/lexicrawler
    ```

2.  **Navigate to the project directory:**

    ```bash
    cd $GOPATH/src/github.com/h2210316651/lexicrawler # Or wherever your Go projects are located
    ```

3.  **Run the API:**

    ```bash
    go run main.go
    ```

    LexiCrawler API will now be running at `http://localhost:3000`.

### Basic Usage

Send a GET request to the `/crawl` endpoint with the `url` query parameter to crawl a specific webpage and receive its Markdown content:

```bash
curl "http://localhost:3000/crawl?url=https://www.example.com"
```

**Example Response:**

```markdown
# Example Domain

This domain is for use in illustrative examples in documents. You may use this
    domain in literature without prior coordination or asking for permission.

[More information...](http://www.iana.org/domains/example)
```

---

## ⚙️ Configuration - Fine-Tune Your Crawls

LexiCrawler offers several configuration options to tailor its behavior. You can control these via:

*   **API Query Parameters:** For on-the-fly adjustments per request.
*   **Modifying `CrawlerConfig` in `main.go`:** For setting default crawler behaviors.

### API Query Parameters

| Parameter        | Description                                                                 | Type    | Default     |
|-----------------|-----------------------------------------------------------------------------|---------|-------------|
| `url`            | **Required.** The URL to crawl.                                             | String  | -           |
| `readability`    | Enable/disable readability enhancement.                                   | Boolean | `false`     |
| `js`             | Enable/disable JavaScript rendering (dynamic content handling).              | Boolean | `false`     |
| `screenshots`    | Enable/disable screenshot capture.                                        | Boolean | `false`     |
| `cache`          | Enable/disable content caching.                                           | Boolean | `false`     |
| `heuristics`     | Enable/disable basic heuristics filtering.                                | Boolean | `false`     |
| `content_selectors` | Comma-separated CSS selectors to target specific content sections.         | String  | (Full page) |


**Example API Request with Parameters:**

```bash
curl "http://localhost:3000/crawl?url=https://blog.example.com/article-title&readability=true&js=true&screenshots=false"
```

### `CrawlerConfig` Options (in `main.go`)

```go
config := CrawlerConfig{
    StartURL:        "", // Set via API parameter
    AllowedDomains:  []string{}, // Dynamically set from URL
    MaxDepth:        2,        // Default crawl depth
    EnableJS:        false,    // Default JS rendering off
    EnableScreenshots: false, // Default screenshots off
    CacheEnabled:    false,    // Default caching off
    HeuristicsEnabled: false, // Default heuristics off
    EnableReadability: false, // Default readability off
    // ContentSelectors: []string{}, // Can be set here or via API parameter
}
```

Modify these values in `main.go` to set the default behavior of your crawler.  API query parameters will override these defaults for individual requests.

---

## 📚 Usage Examples -  Unlocking Web Content for LLMs

Here are a few examples to illustrate LexiCrawler's versatility:

**1. Crawling a Blog Post with Readability and Markdown Output:**

```bash
curl "http://localhost:3000/crawl?url=https://example-blog.com/great-article&readability=true"
```

**Sample Markdown Output (cleaned and readable):**

```markdown
## The Greatness of Example Blog Articles

This is the main content of a fantastic blog article...

... more insightful paragraphs ...

**Key Takeaways:**

*   Point 1
*   Point 2
*   Point 3

[Read the full article on Example Blog](https://example-blog.com/great-article)
```

**2. Crawling a Dynamic Web Application with JavaScript Rendering:**

```bash
curl "http://localhost:3000/crawl?url=https://dynamic-webapp.com/dashboard&js=true"
```

LexiCrawler will use `chromedp` to render the page, ensuring content loaded by JavaScript is captured.

**3. Targeting Specific Content Sections with CSS Selectors:**

Let's say you only want to extract the main article body from a news website, identified by the CSS class `.article-body`:

```bash
curl "http://localhost:3000/crawl?url=https://news-site.com/latest-news&content_selectors=.article-body"
```

LexiCrawler will only process and return the Markdown content found within elements matching the `.article-body` selector.

---

## 🤝 Contributing - Build the Future of LLM Data

LexiCrawler is open source and thrives on community contributions!  We welcome:

*   **Feature Requests:**  Have a great idea to enhance LexiCrawler? Open an issue!
*   **Bug Reports:**  Found a bug? Please report it with clear steps to reproduce.
*   **Pull Requests:**  Code contributions are highly appreciated!  Please follow these guidelines:
    *   **Code Style:**  Adhere to standard Go coding conventions.
    *   **Testing:**  Include tests for new features or bug fixes whenever possible.
    *   **Clear Commit Messages:**  Write descriptive commit messages explaining your changes.

To contribute, please fork the repository, make your changes in a branch, and submit a pull request.

---

## 📜 License

LexiCrawler is released under the [MIT License](LICENSE).  Feel free to use, modify, and distribute it as you wish.

---

##  Author & Maintainer (Optional)

[Srinath Pulaverthi](https://github.com/h2210316651)



**Let's make web data truly LLM-ready together!  Happy crawling!**