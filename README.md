<div align="center">

# DuckCloud

<h3 align='center'> A self-hostable personal cloud accessible for everyone. </h3>

DuckCloud is an open-source software that allows you to keep all your documents safe at home
and access them from everywhere via internet.<br/>

<p align="center">
    <a href="https://github.com/theduckcompany/duckcloud/commits/master">
    <img src="https://img.shields.io/github/last-commit/theduckcompany/duckcloud.svg?style=flat-square&logo=github&logoColor=white"
         alt="GitHub last commit">
    <a href="https://github.com/theduckcompany/duckcloud/issues">
    <img src="https://img.shields.io/github/issues-raw/theduckcompany/duckcloud.svg?style=flat-square&logo=github&logoColor=white"
         alt="GitHub issues">
    <a href="https://github.com/theduckcompany/duckcloud/pulls">
    <img src="https://img.shields.io/github/issues-pr-raw/theduckcompany/duckcloud.svg?style=flat-square&logo=github&logoColor=white"
         alt="GitHub pull requests">
</p>
      
<p align="center">
  <a href="https://docs.duckcloud.fr">Documentation</a> •
  <a href="https://docs.duckcloud.fr/credits/">Credits</a>
</p>

</div>


> [!WARNING]  
> This project is in alpha stage. Things are moving fast and some occasional breaking changes could appears.

# What is it ?

DuckCloud is a self-hostable open-source file hosting service. It provides functionalities similar to [Dropbox][1],
[Microsoft 365][2] or [Google Drive][3] but is concieved to run on low cost servers like a Raspberry Pi for for
a family.

#### Accessible for everyone

DuckCloud aims to be easiest open-source alternative for file hosting and strive to stay as simple as possible. 
It is also accessible for screen readers and color impaired peoples.

#### Secure by default

DuckcCloud incorporate the state-of-the-art security standards and best practices to ensure your data
protection. All the files are automatically and transparently encrypted on disk by default.

#### Easy to install 

The installation aims to be as easy as possible. Grandma should be able install it (at least we try).

# Why should I use it ?

If you care about your data, if you are tired of BigCorps stealing and selling your family data, you
can find a solution in Duckcloud:

#### It's open-source. 

The source code is public and can be modified, copied, or redistributed by anyone. This is important not 
only for the developers wanting to change the source code but for the users too. With a community of 
open-source developer a product ensure itself a more stable and perennial support.

#### It's self-hostable.

At home on your own hardware or in the cloud you trust, DuckCloud have an easy installation and its maintenance 
process.




## Features / Roadmap
- [x] A virtual file system with a file deduplication system and a data at rest encryption
- [x] A WebDAV integration to connect all your WebDAV compliant devices 
- [x] A web interface to interact with you files.
- [x] A web interface for managing the users, settings and navigate the files
- [ ] A contact registry with a CardDAV integration and a web interface
- [ ] An event registry with a CalDAV integration and a web interface
- [ ] A backup service with end-to-end encryption available with a few clicks

## Installation

Please check the [documentation](https://docs.duckcloud.fr/installation-guide/introduction/)

[1]: https://en.wikipedia.org/wiki/Dropbox
[2]: https://en.wikipedia.org/wiki/Microsoft_365
[3]: https://en.wikipedia.org/wiki/Google_Drive
