var app = angular.module('huntApp', ['ngMaterial', 'ngCookies']);

app.config(function($mdThemingProvider) {
	$mdThemingProvider.theme('default')
	    .primaryPalette('blue')
	    .accentPalette('indigo');
    });

app.factory('api', function($http) {
	var result = {};
	var listeners = [];
	var channel;
	var socket;
	
	result.getURL = function (path) {
	    var result = "/api/" + path + "?hunt_id=" + huntId;
	    if (isAdmin) {
		result = "/admin" + result;
	    }
	    return result
	}

	result.getState = function() {
	    return $http.get(result.getURL("state"));
	}

	result.getPuzzles = function() {
	    return $http.get(result.getURL("puzzles"));
	}

	result.getIngredients = function() {
	    return $http.get(result.getURL("ingredients"));
	}

	result.setIngredients = function(newIngredients) {
	    return $http.get(result.getURL("updateingredients") +
			     "&ingredients=" + encodeURIComponent(newIngredients));
	}

	result.getTeams = function() {
	    return $http.get(result.getURL("teams"));
	}

	result.addTeam = function(name, password, novice) {
	    return $http.get(result.getURL("addteam") + 
			     "&name=" + encodeURIComponent(name) +
			     "&password=" + encodeURIComponent(password) +
			     "&novice=" + novice);
	}

	result.deleteTeam = function(id) {
	    return $http.get(result.getURL("deleteteam") +
			     "&team_id=" + id);
	}

	result.getLeaderboard = function() {
	    return $http.get(result.getURL("leaderboard"));
	}

	result.getLeaderboardUpdate = function(id) {
	    return $http.get(result.getURL("leaderboardupdate") +
			     "&puzzleid=" + id);
	}

	result.getConsole = function(id) {
	    return $http.get(result.getURL("console"));
	}

	result.getTeamInfo = function() {
	    return $http.get(result.getURL("teaminfo"));
	}

	result.submitAnswer = function(pid, answer) {
	    return $http.get(result.getURL("submitanswer") +
			     "&puzzleid=" + pid +
			     "&answer=" + encodeURIComponent(answer));
	}

	result.advanceState = function(currentState) {
	    return $http.get(result.getURL("advancestate") +
			     "&currentstate=" + currentState);
	}

	result.updatePuzzle = function(id, name, answer) {
	    return $http.get(result.getURL("updatepuzzle") +
			     "&puzzleid=" + id +
			     "&name=" + encodeURIComponent(name) +
			     "&answer=" + encodeURIComponent(answer));
	}

	result.getSurveyInfo = function() {
	    return $http.get(result.getURL("survey"));
	}

	result.getAdminSurveyInfo = function() {
	    return $http.get(result.getURL("adminsurvey"));
	}

	result.submitSurvey = function(res) {
	    return $http.get(result.getURL("submitsurvey") +
			     "&result=" + res);
	}

	result.getFinalScores = function() {
	    return $http.get(result.getURL("finalscores"));
	}
	
	result.addListener = function(listener) {
	    listeners.push(listener);
	}

	result.onMessage = function(message) {
	    var j = JSON.parse(message.data);
	    if (j.K == "refresh") {
		for (var i = 0; i < listeners.length; i++) {
		    if (listeners[i].hasOwnProperty("refresh")) {
			listeners[i].refresh();
		    }
		}
	    } else {
		for (var i = 0; i < listeners.length; i++) {
		    if (listeners[i].hasOwnProperty("onMessage")) {
			listeners[i].onMessage(j);
		    }
		}
	    }
	}

	result.onError = function(message) {
	    console.log("Channel error");
	    console.log(message);
	}

	result.openChannel = function() {
	    var doRefresh = false;
	    if (result.socket != null) {
		result.socket.close();
		result.secket = null;
		doRefresh = true;
	    }
	    $http.get(result.getURL("channel")).success(function (response) {
		    channel = new goog.appengine.Channel(response);
		    result.socket = channel.open();
		    result.socket.onmessage = result.onMessage;
		    result.socket.onerror = result.onError;
		    // Refresh the channel in 90 min.
		    window.setTimeout(function () {
			    result.openChannel();
			}, 1000 * 60 * 90);
		    if (doRefresh) {
			for (var i = 0; i < listeners.length; i++) {
			    if (listeners[i].hasOwnProperty("refresh")) {
				listeners[i].refresh();
			    }
			}
		    }
		});
	}

	result.socket = null
	result.openChannel();
	return result;
    });

app.controller('consoleCtrl', function($scope, api) {
	$scope.onMessage = function(message) {
	    if (message.K == "consoleupdate") {
		$scope.$apply(function() {
			$scope.Lines.unshift(message.V);
		    });
	    }
	}

	$scope.Lines = [];
	api.getConsole().success(function (response) {
		if (response != null) {
		    $scope.Lines = response;
		}
	    });
	api.addListener($scope);
    });

app.controller('leaderboardCtrl', function ($scope, $mdToast, api) {
	$scope.refresh = function() {
	    api.getLeaderboard().success(function (response) {
		    $scope.Leaderboard = response;
		});
	}

	$scope.updatePuzzle = function(id) {
	    api.getLeaderboardUpdate(id).success(function (response) {
		    for (var i = 0; i < $scope.Leaderboard.Progress.length; i++) {
			if (id == $scope.Leaderboard.Progress[i].ID) {
			    $scope.Leaderboard.Progress[i].Updatable = response;
			}
		    }
		});
	}

	$scope.submitAnswer = function(pid, answer) {
	    api.submitAnswer(pid, answer).success(function (response) {
		    window.alert(response);
		});
	}

	$scope.puzzleFormat = {
	    false: "Non-paper",
	    true: "Paper",
	}

	$scope.onMessage = function(message) {
	    if (message.K == "leaderboardupdate") {
		$scope.updatePuzzle(message.V);
	    }
	}

	api.addListener($scope);
	$scope.refresh();
    });

app.controller('puzzlesCtrl', function ($scope, api) {
	$scope.refresh = function() {
	    api.getPuzzles().success(function (response) {
		    $scope.puzzles = response;
		    $scope.hasPuzzles = (response != null && response.Puzzles != null && response.Puzzles.length > 0);
		});
	}

	$scope.onMessage = function(message) {
	    if (message.K == "puzzlesupdate") {
		$scope.refresh();
	    }
	}

	$scope.updatePuzzle = function(id, name, answer) {
	    api.updatePuzzle(id, name, answer);
	}

	$scope.hasPuzzles = false;
	api.addListener($scope);
	$scope.refresh();
    });

app.controller('ingredientsCtrl', function ($scope, api) {
	$scope.refresh = function() {
	    api.getIngredients().success(function (response) {
		    $scope.ingredients = response;
		});
	}

	$scope.onMessage = function(message) {
	    if (message.K == "ingredientsupdate") {
		$scope.refresh();
	    }
	}

	$scope.updateIngredients = function() {
	    api.setIngredients($scope.newIngredients).success(function () {
		    $scope.newIngredients = "";
		});
	}

	$scope.newIngredents = "";
	api.addListener($scope);
	$scope.refresh();
    });

app.controller('teamsCtrl', function ($scope, api) {
	$scope.refresh = function() {
	    api.getTeams().success(function (response) {
		    $scope.teams = response;
		});
	}

	$scope.onMessage = function(message) {
	    if (message.K == "teamsupdate") {
		$scope.refresh();
	    }
	}

	$scope.addTeam = function() {
	    api.addTeam($scope.newTeamName, $scope.newTeamPassword,
			$scope.newTeamNovice).success(function () {
				$scope.newTeamName = "";
				$scope.newTeamPassword = "";
				$scope.newTeamNovice = false;
			    });
	}

	$scope.deleteTeam = function(id) {
	    api.deleteTeam(id);
	}

	$scope.newTeamName = "";
	$scope.newTeamPassword = "";
	$scope.newTeamNovice = false;
	api.addListener($scope);
	$scope.refresh();
    });

app.controller('stateCtrl', function ($scope, $http, api) {
	$scope.refresh = function() {
	    api.getState().success(function (response) {
		    $scope.state = response;
		    $scope.advanceable = ($scope.state != 6);
		});
	}
	
	$scope.advanceState = function() {
	    r = window.confirm("Are you sure you wish to advance the state? This cannot be undone.");
	    if (r) {
		api.advanceState($scope.state);
	    }
	}

	$scope.stateTable = {
	    0: "Pre-launch",
	    1: "Novice ingredients released",
	    2: "Ingredients released",
	    3: "Solving",
	    4: "Surveying",
	    5: "Tallying done",
	    6: "Results released",
	}

	$scope.advanceable = false;
	api.addListener($scope);
	$scope.refresh();
    });

app.controller('signinCtrl', function($scope, $cookies, $mdToast, api) {
	$scope.refresh = function() {
	    api.getTeamInfo().success(function (response) {
		    $scope.teamInfo = response;
		    if (response.BadSignIn) {
			$mdToast.show($mdToast.simple().content("Bad password"));
			$scope.logout()
		    }
		});
	}

	$scope.login = function() {
	    $cookies.put("team_id", $scope.teamID);
	    $cookies.put("password", $scope.password);
	    $scope.teamID = "";
	    $scope.password = "";
	    $scope.refresh();
	    api.openChannel();
	}

	$scope.logout = function() {
	    $cookies.remove("team_id");
	    $cookies.remove("password");
	    $scope.refresh();
	    api.openChannel();
	}

	$scope.refresh();
    });

app.controller('adminSurveyCtrl', function($scope, api) {
	$scope.refresh = function() {
	    api.getAdminSurveyInfo().success(function (response) {
		    $scope.adminSurveyInfo = response;
		});
	}

	$scope.onMessage = function(message) {
	    if (message.K == "surveyupdate") {
		$scope.refresh();
	    }
	}

	api.addListener($scope);
	$scope.refresh();
    });

app.controller('surveyCtrl', function($scope, api) {
	$scope.refresh = function() {
	    api.getSurveyInfo().success(function (response) {
		    $scope.surveyInfo = response;
		    if ($scope.surveyInfo.Puzzles != null) {
			for (var i = 0; i < $scope.surveyInfo.Puzzles.length; i++) {
			    var p = $scope.surveyInfo.Puzzles[i];
			    p.fun = 1;
			    p.presentation = 1;
			    p.ingredients = 1;
			}
		    }
		});
	}

	$scope.submit = function() {
	    var result = "";
	    for (var i = 0; i < $scope.surveyInfo.Puzzles.length; i++) {
		var p = $scope.surveyInfo.Puzzles[i];
		result = result.concat(p.fun);
		result = result.concat(p.presentation);
		result = result.concat(p.ingredients);
	    }
	    api.submitSurvey(result);
	}

	$scope.onMessage = function(message) {
	    if (message.K == "surveyupdate") {
		$scope.refresh();
	    }
	}

	api.addListener($scope);
	$scope.refresh();
    });

app.controller('finalScoresCtrl', function($scope, api) {
	$scope.refresh = function() {
	    api.getFinalScores().success(function (response) {
		    $scope.finalScores = response;
		});
	}

	api.addListener($scope);
	$scope.refresh();
    });