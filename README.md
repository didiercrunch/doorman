#  Doorman [![Build Status](https://travis-ci.org/didiercrunch/doorman.svg)](https://travis-ci.org/didiercrunch/doorman)

doorman client.  For the doorman server see [doorman server](https://github.com/didiercrunch/doorman server)


## what it is?

This library can be use in two ways; A/B testing and feature gating.

### A/B testing

The best way to explain *A/B testing* is via an example.  Imagine that you have a button in you blog
and that you want to set its colour to maximaze the number of viewers that click on it.  For simplicity,
let's assume the colour can be red or magenta.

*A/B* testing allows you to randomly select a colour for the button depending on your viewers.  For example,
you could deceide to show to 10% of your users the magenta button while 90% of your user will see the button
red.

Then, you can collect data on your user choice (feature not include in this library) and make a data supported
decision to improved you blog.


### Feature gating

The best way to explain *feature gating* is to give an example.  Imaging that you are using *company A* to
manage the discussion section of your blog and that you want to switch to *company B*.  You would like to
make the move but you are scared that the switch will cause failure in unexpected part of your blog.

What you can do is to change graduly from *company A* to *company B*.  At first, you set 1% of your
viewer to see *company A*, if no issue occure, you move to 5% then to 10% and up to 100%.

## example

There is an example at [example/main.go](https://github.com/didiercrunch/doorman/blob/master/example/main.go) that
change randomly the colour of the background by the url path.
