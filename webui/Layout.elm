module Layout exposing (header)

import Types
import Html exposing (Html, div, span, text, img)
import Html.Attributes exposing (src)
import Bulma.CDN exposing (..)
import Bulma.Components exposing (..)
import Bulma.Modifiers exposing (..)


header : Html Types.Msg
header =
    div []
        [ stylesheet
        , ptNavbar
        ]


ptNavbarBurger : Html Types.Msg
ptNavbarBurger =
    navbarBurger False
        []
        [ span [] []
        , span [] []
        , span [] []
        ]


ptNavbar : Html Types.Msg
ptNavbar =
    navbar { color = Dark, transparent = False }
        []
        [ navbarBrand []
            ptNavbarBurger
            [ navbarItem False [] [ text "Papertrail" ] ]
        , navbarMenu False
            []
            [ navbarStart []
                [ navbarItemLink False [] [ text "Home" ]
                , navbarItemLink False [] [ text "Rawr" ]
                ]
            , navbarEnd []                []
            ]
        ]
