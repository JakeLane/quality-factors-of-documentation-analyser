library("jsonlite")
library("Hmisc")

data <- do.call(rbind, 
                     lapply(paste(readLines(file("data.json"), warn=FALSE),
                                  collapse=""), 
                            jsonlite::fromJSON))

linearRegressionTotalBytes <- lm(data$TotalBytes~data$Forks)
linearRegressionGunningFogIndex <- lm(data$GunningFogIndex~data$Forks)

par(mfrow=c(1, 1))
scatter.smooth(
  x=data$TotalBytes,
  y=data$Forks,
  xlab="Size of documentation in bytes",
  ylab="Number of forks"
)
abline(linearRegressionTotalBytes, col="red") # regression line (y~x) 

scatter.smooth(
  x=data$GunningFogIndex,
  y=data$Forks,
  xlab="Gunning Fog Index",
  ylab="Number of forks"
)
abline(linearRegressionGunningFogIndex, col="red") # regression line (y~x) 

par(mfrow=c(1, 3))
boxplot(data$TotalBytes, main="Size of documentation")
boxplot(data$GunningFogIndex, main="Guning Fog Index")
boxplot(data$Forks, main="Number of Forks")

library(e1071)
par(mfrow=c(1, 3))  # divide graph area in 2 columns
plot(density(data$TotalBytes), main="Density Plot: Size of documentation", ylab="Frequency", sub=paste("Skewness:", round(e1071::skewness(data$TotalBytes), 2)))  # density plot for 'speed'
polygon(density(data$TotalBytes), col="red")
plot(density(data$GunningFogIndex), main="Density Plot: Gunning Fog Index", ylab="Frequency", sub=paste("Skewness:", round(e1071::skewness(data$GunningFogIndex), 2)))  # density plot for 'dist'
polygon(density(data$GunningFogIndex), col="red")
plot(density(data$Forks), main="Density Plot: Number of Forks", ylab="Frequency", sub=paste("Skewness:", round(e1071::skewness(data$Forks), 2)))  # density plot for 'dist'
polygon(density(data$Forks), col="red")

pearsonTotalBytes <- cor(data[,"Forks"], data[,"TotalBytes"], method="pearson")
pearsonGunningFogIndex <- cor(data[,"Forks"], data[,"GunningFogIndex"], method="pearson")
