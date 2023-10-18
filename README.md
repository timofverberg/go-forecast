
1. 
    Creating a .env file is required.  
        echo 'USERAGENT=Go/ForecastProject  
        GEOCODEAPIURL=https://geocode.maps.co/search  
        FORECASTAPIURL=https://api.met.no/weatherapi/locationforecast/2.0/compact' > .env  

    
2. Compiling  
    go build


3. Usage  
    ./weather "Stockholm, Stockholms kommun, Stockholm County, 111 29, Sweden"





