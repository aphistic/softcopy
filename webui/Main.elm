module Main exposing (..)

import Bulma.Columns exposing (..)
import Bulma.Form exposing (..)
import Bulma.Layout exposing (..)
import Bulma.Modifiers exposing (..)
import Html exposing (Html, text)
import Html.Attributes exposing(placeholder)
import Layout
import Types


type alias Model =
    {}


main : Program Never Model Types.Msg
main =
    Html.beginnerProgram
        { model = {}
        , view = view
        , update = \msg -> \model -> model
        }


view : Model -> Html Types.Msg
view model =
    let
        mainModifiers =
            { offset = Auto
            , widths =
                { mobile = Just Width8
                , tablet = Just Width8
                , desktop = Just Width8
                , widescreen = Just Width8
                , fullHD = Just Width8
                }
            }
    in
        Html.main_ []
            [ container []
                [ Layout.header
                , container []
                    [ columns columnsModifiers
                        []
                        [ column mainModifiers
                            []
                            [ controlText controlInputModifiers [] [ placeholder "Search" ] []
                            ]
                        , column columnModifiers [] [ text "Sidebar" ]
                        ]
                    ]
                ]
            ]
