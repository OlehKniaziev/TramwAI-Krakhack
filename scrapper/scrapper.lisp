(defpackage scrapper
  (:use "COMMON-LISP"))

(defvar *url* "https://www.krakow.pl/kalendarium/1919,artykul,kalendarium.html")
(defvar *request* (dex:get *url*))
(defvar *parsed-content* (plump:parse *request*))
