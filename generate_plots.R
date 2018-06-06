library("jsonlite")
library("Hmisc")

data <- flatten(do.call(rbind, 
                     lapply(paste(readLines(file("output-v1.json"), warn=FALSE),
                                  collapse=""), 
                            jsonlite::fromJSON)))

plot(
  x=data[,"RepoInfo.forks_count"],
  y=data[,"TotalBytes"],
  xlab="Number of forks (logarithmic scale)",
  ylab="Size of documentation in bytes (logarithmic scale)",
  log="xy"
)

pearson <- cor(data[,"RepoInfo.forks_count"], data[,"TotalBytes"], method="pearson")